package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"
	"time"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/services/parsers"
	"gorm.io/gorm"
)

type ReceiptIngestionService interface {
	IngestXML(file *multipart.FileHeader, user *models.User) (*models.InboxEntry, error)
	IngestXMLBytes(raw []byte, user *models.User) (*models.InboxEntry, error)
	IngestURL(rawURL string, user *models.User) (*models.InboxEntry, error)
}

type receiptIngestionService struct {
	db         *gorm.DB
	inbox      InboxService
	xmlParser  parsers.ReceiptXMLParser
	htmlParser parsers.ReceiptHTMLParser
	httpClient *http.Client
}

func isSafeIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return false
	}
	return true
}

func secureTransport() *http.Transport {
	dialer := &net.Dialer{
		Timeout:   20 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				if !isSafeIP(ip) {
					return nil, fmt.Errorf("SSRF prevention: IP %s is not allowed", ip.String())
				}
			}
			// Connect directly to the resolved IP to prevent DNS rebinding
			return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func NewReceiptIngestionService(db *gorm.DB, inbox InboxService) ReceiptIngestionService {
	return &receiptIngestionService{
		db:         db,
		inbox:      inbox,
		xmlParser:  parsers.NewSilpoXMLParser(),
		htmlParser: parsers.NewSilpoHTMLParser(),
		httpClient: &http.Client{
			Timeout:   20 * time.Second,
			Transport: secureTransport(),
		},
	}
}

func (s *receiptIngestionService) IngestXML(file *multipart.FileHeader, user *models.User) (*models.InboxEntry, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	raw, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read xml file: %w", err)
	}

	return s.IngestXMLBytes(raw, user)
}

func (s *receiptIngestionService) IngestXMLBytes(raw []byte, user *models.User) (*models.InboxEntry, error) {
	parsed, err := s.xmlParser.Parse(raw)
	if err != nil {
		return nil, err
	}
	parsed.SourceType = models.ReceiptSourceTypeXML
	parsed.RawSource = string(raw)

	return s.createInboxFromParsedReceipt(parsed, models.ReceiptOriginTelegramXML, "", user)
}

func (s *receiptIngestionService) IngestURL(rawURL string, user *models.User) (*models.InboxEntry, error) {
	parsedURL, err := neturl.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, errors.New("invalid receipt url")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, errors.New("unsupported url scheme")
	}

	body, contentType, err := s.fetchURLBody(parsedURL.String())
	if err != nil {
		return nil, err
	}

	receipt, err := s.parseReceiptFromURLBody(body, contentType, parsedURL.String())
	if err != nil {
		return nil, err
	}
	receipt.SourceType = models.ReceiptSourceTypeURL
	receipt.RawSource = string(body)

	return s.createInboxFromParsedReceipt(receipt, models.ReceiptOriginTelegramURL, parsedURL.String(), user)
}

func (s *receiptIngestionService) fetchURLBody(rawURL string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "WeGaS-Finance/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch receipt url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("receipt url returned status %d", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, 2*1024*1024)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read receipt response: %w", err)
	}

	return body, resp.Header.Get("Content-Type"), nil
}

func (s *receiptIngestionService) parseReceiptFromURLBody(body []byte, contentType string, sourceURL string) (*parsers.ParsedReceipt, error) {
	trimmed := bytes.TrimSpace(body)

	if strings.Contains(strings.ToLower(contentType), "xml") || bytes.HasPrefix(trimmed, []byte("<?xml")) || bytes.HasPrefix(trimmed, []byte("<RQ")) {
		return s.parseReceiptXMLWithMerchantFallback(body, sourceURL)
	}

	if strings.Contains(strings.ToLower(contentType), "html") || bytes.Contains(trimmed, []byte(`cheque-goods`)) {
		receipt, err := s.htmlParser.Parse(body)
		if err == nil {
			return receipt, nil
		}
	}

	xmlBlock := extractEmbeddedXML(body)
	if len(xmlBlock) > 0 {
		return s.parseReceiptXMLWithMerchantFallback(xmlBlock, sourceURL)
	}

	xmlLink := extractXMLLink(body, sourceURL)
	if xmlLink != "" {
		xmlBody, _, err := s.fetchURLBody(xmlLink)
		if err != nil {
			return nil, err
		}
		return s.parseReceiptXMLWithMerchantFallback(xmlBody, sourceURL)
	}

	return nil, errors.New("unsupported receipt url format")
}

func extractEmbeddedXML(body []byte) []byte {
	content := string(body)
	start := strings.Index(content, "<?xml")
	if start == -1 {
		start = strings.Index(content, "<RQ")
	}
	end := strings.LastIndex(content, "</RQ>")
	if start == -1 || end == -1 || end <= start {
		return nil
	}
	return []byte(content[start : end+len("</RQ>")])
}

func extractXMLLink(body []byte, sourceURL string) string {
	content := string(body)
	re := regexp.MustCompile(`https?://[^\s"'<>]+\.xml`)
	if match := re.FindString(content); match != "" {
		return match
	}

	reHref := regexp.MustCompile(`href=["']([^"']+\.xml)["']`)
	matches := reHref.FindStringSubmatch(content)
	if len(matches) < 2 {
		return ""
	}

	base, err := neturl.Parse(sourceURL)
	if err != nil {
		return matches[1]
	}
	ref, err := neturl.Parse(matches[1])
	if err != nil {
		return matches[1]
	}
	return base.ResolveReference(ref).String()
}
