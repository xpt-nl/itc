package report

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var version = 2.1
var salesEndpoint = "https://reportingitc-reporter-sh-prz.apple.com/reportservice/sales/v1"
var financeEndpoint = "https://reportingitc-reporter-sh-prz.apple.com/reportservice/finance/v1"

// Client is reporter client
type Client struct {
	AccessToken string
	Mode        string
	httpCli     *http.Client
}

// Config base properties
type Config struct {
	AccessToken string
	Mode        string
}

// Request to Reporter endpoints
type Request struct {
	AccessToken string `json:"accesstoken"`
	Account     string `json:"account"`
	Version     string `json:"version"`
	Mode        string `json:"mode"`
	SalesURL    string `json:"salesurl"`
	FinanceURL  string `json:"financeurl"`
	QueryInput  string `json:"queryInput"`
}

// SetAccount sets account as a query escape string
func (r *Request) SetAccount(account int) {
	r.Account = url.QueryEscape(strconv.Itoa(account))
}

// NewClient yield a new Client instance
func NewClient(cfg Config) (*Client, error) {
	err := cfg.Check()
	if err != nil {
		return nil, err
	}
	return &Client{
		AccessToken: cfg.AccessToken,
		Mode:        cfg.Mode,
		httpCli: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 180 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 300 * time.Second,
				DisableCompression:    true,
				DisableKeepAlives:     false,
			},
		},
	}, nil
}

func (c Client) CloseIdleConnections() {
	c.httpCli.CloseIdleConnections()
}

// GetSalesStatus return Sales.getStatus response
func (c Client) GetSalesStatus(ctx context.Context) ([]byte, error) {
	req := c.getBaseRequest()
	req.QueryInput = "%5Bp%3DReporter.properties%2C+Sales.getStatus%5D"
	return c.send(ctx, salesEndpoint, req)
}

// GetFinanceStatus return Finance.getStatus response
func (c Client) GetFinanceStatus(ctx context.Context) ([]byte, error) {
	req := c.getBaseRequest()
	req.QueryInput = "%5Bp%3DReporter.properties%2C+Finance.getStatus%5D"
	return c.send(ctx, financeEndpoint, req)
}

// GetSalesAccounts return Sales.getAccounts response
func (c Client) GetSalesAccounts(ctx context.Context) ([]byte, error) {
	req := c.getBaseRequest()
	req.QueryInput = "%5Bp%3DReporter.properties%2C+Sales.getAccounts%5D"
	return c.send(ctx, salesEndpoint, req)
}

// GetFinanceAccounts return Finance.getAccounts response
func (c Client) GetFinanceAccounts(ctx context.Context) ([]byte, error) {
	req := c.getBaseRequest()
	req.QueryInput = "%5Bp%3DReporter.properties%2C+Finance.getAccounts%5D"
	return c.send(ctx, financeEndpoint, req)
}

// GetSalesVendors return Sales.getVendors response
func (c Client) GetSalesVendors(ctx context.Context, account int) ([]byte, error) {
	if account <= 0 {
		return nil, errors.New("wrong account number")
	}
	req := c.getBaseRequest()
	req.QueryInput = fmt.Sprintf("%%5Bp%%3DReporter.properties%%2C+a%%3D%d%%2C+Sales.getVendors%%5D", account)
	return c.send(ctx, salesEndpoint, req)
}

// GetFinanceVendorsAndRegions return Finance.getVendors response
func (c Client) GetFinanceVendorsAndRegions(ctx context.Context, account int) ([]byte, error) {
	if account <= 0 {
		return nil, errors.New("wrong account number")
	}
	req := c.getBaseRequest()
	req.SetAccount(account)
	req.QueryInput = fmt.Sprintf("%%5Bp%%3DReporter.properties%%2C+m%%3D%s%%2C+Finance.getVendorsAndRegions%%5D", c.Mode)
	return c.send(ctx, financeEndpoint, req)
}

// GetSalesReport return Sales.getReport response (is report file or error)
func (c Client) GetSalesReport(ctx context.Context, account, vendor int, reportType, reportSubType, dateType, date string) ([]byte, error) {
	err := validateSalesReportArgs(account, vendor, reportType, reportSubType, dateType, date)
	if err != nil {
		return nil, err
	}
	req := c.getBaseRequest()
	req.SetAccount(account)
	qI := "%%5Bp%%3DReporter.properties%%2C+m%%3D%s%%2C+Sales.getReport%%2C+%d%%2C%s%%2C%s%%2C%s%%2C%s%%5D"
	req.QueryInput = fmt.Sprintf(qI, c.Mode, vendor, reportType, reportSubType, dateType, date)
	return c.send(ctx, salesEndpoint, req)
}

// GetFinanceReport return Finance.getReport response (is report file or error)
func (c Client) GetFinanceReport(ctx context.Context, account, vendor int, regionCode, reportType string, fiscalYear, fiscalPeriod int) ([]byte, error) {
	err := validateFinancialReportArgs(account, vendor, regionCode, reportType, fiscalYear, fiscalPeriod)
	if err != nil {
		return nil, err
	}
	req := c.getBaseRequest()
	req.SetAccount(account)
	qI := "%%5Bp%%3DReporter.properties%%2C+m%%3D%s%%2C+Finance.getReport%%2C+%d%%2C%s%%2C%s%%2C%d%%2C%d%%5D"
	req.QueryInput = fmt.Sprintf(qI, c.Mode, vendor, regionCode, reportType, fiscalYear, fiscalPeriod)
	return c.send(ctx, financeEndpoint, req)
}

func (c Client) send(ctx context.Context, endpoint string, r Request) ([]byte, error) {
	q, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("jsonRequest=%s", string(q))
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(query))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html, image/gif, image/jpeg, *; q=.2, */*; q=.2")
	req.Header.Add("User-Agent", "Java/1.8.0_112")
	req.Header.Add("Connection", "keep-alive")

	res, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}
	return body, nil
}

func (c Client) getBaseRequest() Request {
	return Request{
		AccessToken: url.QueryEscape(c.AccessToken),
		Version:     url.QueryEscape(fmt.Sprintf("%.1f", version)),
		Mode:        url.QueryEscape(c.Mode),
		SalesURL:    url.QueryEscape(salesEndpoint),
		FinanceURL:  url.QueryEscape(financeEndpoint),
	}
}

func (cfg Config) Check() error {
	if cfg.Mode != "Normal" && cfg.Mode != "Robot.xml" {
		return errors.New("undefined mode. Use available modes: Normal or Robot.xml")
	}
	if cfg.AccessToken == "" {
		return errors.New("access token not set")
	}
	return nil
}

func validateSalesReportArgs(account, vendor int, reportType, reportSubType, dateType, date string) error {
	if account <= 0 {
		return errors.New("wrong account number")
	}
	if vendor <= 0 {
		return errors.New("wrong vendor number")
	}

	if reportType != "Sales" && reportType != "Newsstand" {
		return errors.New("wrong ReportType, use: Sales or Newsstand")
	}

	switch reportSubType {
	case "Summary":
	case "Detailed":
	case "Opt-In":
	default:
		return errors.New("wrong ReportSubType, use: Summary, Detailed or Opt-In")
	}

	switch dateType {
	case "Daily":
		if len(date) != 8 {
			return errors.New("wrong DateType format for Daily Report, use: YYYYMMDD")
		}
	case "Weekly":
		if len(date) != 8 {
			return errors.New("wrong DateType format for Weekly Report, use: YYYYMMDD")
		}
	case "Monthly":
		if len(date) != 6 {
			return errors.New("wrong DateType format for Monthly Report, use: YYYYMM")
		}
	case "Yearly":
		if len(date) != 4 {
			return errors.New("wrong DateType format for Yearly Report, use: YYYY")
		}
	default:
		return errors.New("wrong DateType, use: Daily, Weekly, Monthly or Yearly")
	}

	return nil
}

func validateFinancialReportArgs(account, vendor int, regionCode, reportType string, fiscalYear, fiscalPeriod int) error {
	if account <= 0 {
		return errors.New("wrong account number")
	}
	if vendor <= 0 {
		return errors.New("wrong vendor number")
	}
	if len(regionCode) != 2 {
		return errors.New("wrong region code")
	}
	if reportType != "Financial" {
		return errors.New(`wrong report type: "Currently only one report type is available: Financial"`)
	}
	if fiscalYear > time.Now().Year()+1 || fiscalYear <= 0 {
		return errors.New("wrong fiscal year")
	}
	if fiscalPeriod < 1 || fiscalPeriod > 12 {
		return errors.New("wrong fiscal period, it should be: 1-12")
	}
	return nil
}
