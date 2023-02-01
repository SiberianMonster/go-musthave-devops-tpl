package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/config"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/handlers"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/metrics"
	"github.com/SiberianMonster/go-musthave-devops-tpl/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStoreInterval(t *testing.T) {

	tests := []struct {
		name           string
		storeParameter string
		want           int
	}{
		{
			name:           "trial run",
			storeParameter: "3m",
			want:           3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ParseStoreInterval(&tt.storeParameter)
			assert.Equal(t, tt.want, v)
		})
	}
}

func TestParseRestoreValue(t *testing.T) {

	tests := []struct {
		name    string
		restore string
		want    bool
		wantErr error
	}{
		{
			name:    "trial run",
			restore: "true",
			want:    true,
			wantErr: nil,
		},
		{
			name:    "wrong input",
			restore: "xxx",
			want:    false,
			wantErr: errors.New("could not parse restore value"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, vErr := ParseRestoreValue(&tt.restore)
			assert.Equal(t, tt.want, v)
			assert.Equal(t, tt.wantErr, vErr)
		})
	}
}

func TestParseRsaPrivateKey(t *testing.T) {

	tests := []struct {
		name    string
		pemKey  string
		wantErr error
	}{
		{
			name: "trial run",
			pemKey: "-----BEGIN RSA PRIVATE KEY-----\n" +
				"MIIEowIBAAKCAQEAt35GTCNp+7tkEYEmlSHTZ+Ii1QTZijOZ6AoGXhNQuPAtjUEu" +
				"8+M7SW4RHRdYBD7lLnqoqoM6jUi8b99Hy9h0iYk3k0SKjGmu/OPn22NGOYIFXRKl" +
				"NJ6iNll+CSt543vgNyH5BCCZjDepiqNmCFSAe+55b4a7RD8ruasZ5PNhzot9fRuu" +
				"yzrIKYkD7iisxJZVvzAaO+rVN6dDmXEFC5E4+xEC/fz0rYed8lwLhljnDE1zcHU7" +
				"DSZZCtCXnPRFjU0pWRizkjcbntyIdv4lROULWtRTRjE92+vTeS+yhH16ogARRp+T" +
				"OwxKeHv81tn6UYpuQ9A9/hFrFQG1Ty9VvKcCTwIDAQABAoIBAFL1llbKFBqp6F45" +
				"o/X86xWmmdTxcmEXX1gXYDWcSfyzKgUZGV9OtvlF+BrM+RBCV1+iOSuOVSSXZAq4" +
				"Sj+RR27/SM8eR/2fsmvHpoX75j4N2NrxmRunNPOZlnAS5fLBiOekRm9lRcatS8vQ" +
				"gEr32XcupFyV74i1ftFc2EI4/1lf1wRaiffIhs+yfY6wpWnUwrRDAivVKw9grWF5" +
				"qq9kGeu/qMgrd3SnRuHMPy+KOmei/ugv926ZkbxAQH8JPPAtQmAE4yGiNLqjerg2" +
				"nZ1pzj5OK5+CfvWOlSUAdi0QfKP2nHL5Lj4RxPHuC3IJ18AUXFSASV0uX9a8Qbl7" +
				"O/3KcEECgYEAyWTe3S+MQMWRWp03/cwGecTQixo3TjTCfX0/V3lJTLW+Feg5vsTt" +
				"hxj/xiFEgGcX1alyosQSr5dIZp8sww7j94SfXqv4R3i4QBZdFNQ+2d+rZRIdj2O9" +
				"Sc6yb7968ryiQSePGnFOCfSB1igJyAJZxE8AeWNKuUs14GnfOE8fwV8CgYEA6T7h" +
				"Bpo52cLkLR29CIcnOb8flepTRXJ0RqLG+VWJJI1myxRydyOFB3phiGzKolBI06VB" +
				"UuKMFPNMgCOOJB3d93cQ2R6rbdP96O00uLFrOGjJs8QV6BRt9z2c32XK5Hom4IGU" +
				"YAzFd1jNv1laheRvQ9e1BSHEoR73qp1IU4oNtRECgYEAw/IUtHfSqiKPre5Rz+l2" +
				"U3ueu/ih3sGOibIWsvEa1Dvv2ji8FlRcFpnIIem0UIn9srDPDHZhB97VXqN4VcBj" +
				"JSwwM1h2lHNsMU6Q+fcXv7vTct8RS7XrMaieDAPth8boxyPKJBwhpaXzvX3vJl7D" +
				"IDENcQ2eYnI+1T2tJYg2iVkCgYBhaNs9kKdcZGI63VKW/yrImSMtzuDb/gLFhTGn" +
				"66sM0uj9Ixry2qiyCNA204iE5RalHTz8ypRKI5ntYev49Wg/8z/cDUz23zQJVRdR" +
				"kvb+ZfTm2Jt1gyKxwM+FFNP5O3KFDFjVDEBjqXiz0zNU+6PkJ2/4JrQhvfcdD/am" +
				"vN8goQKBgD/PfolUBXReHeGBepei2DloCVHtCnzrvKbAj79qK4y2isjpJ3F99YXk" +
				"SgEHAZMLlf6Cq0RPbaNRkuo5F3G9e2OcnF8s+PS4LrSpnP1A3WPL0a7X6PaxEBl5" +
				"TCA83Sme6tm8Dazr5QMviaHGKvuXRUTX7TvJKo1m4jdmwRvce6Y6\n" +
				"-----END RSA PRIVATE KEY-----",
			wantErr: nil,
		},
		{
			name:    "wrong input",
			pemKey:  "something",
			wantErr: errors.New("failed to parse PEM block containing the key"),
		},
		{
			name: "wrong encoding",
			pemKey: "-----BEGIN RSA PRIVATE KEY-----\n" +
				"MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCmsoj2W3fLmbOv" +
				"dxTR1anm5mi8nteWXpPtnksHIhriWq0Ysgo7CW4vmwPECPDZJq46Ub9WAJILcrj+" +
				"/ghtZ8Fd6raeCeEb9gPYhtZkCqHyKKc0NY0qHxHEIIlHapM0f44zRdmJtoXcGGlj" +
				"3zurm9b9QIEhxsvhlS80S+GyF38QtM4tWPCJ1pBqaj1wJ3Sg0RzY+pnsarvk01kx" +
				"Rw2XDJNYQPXtOi4dmqzBFf5aYYoo1lO6dKPNnlaYVd9Th4/naeIG0ntYGBefFIip" +
				"UlemhdUc1aqqL/U+JpiAp12ofkKZXe/e63+cc4Vn5g2qpmbtMOG4cbvvBTV8roiQ" +
				"avhZL5BjAgMBAAECggEANDy0UMcfBi1XMoAVhR/4iwPfBGSeWF+w6YB2MHkOhao2" +
				"nguEyzVMUxy3lGHc35+Qb3QYimHJYk8EC9wdVfNyk/SuX13nLfTtBZhTbKwsTY9R" +
				"vjmdz/pGffhYLIoIMSZbsFOONOp+jhcUR5i3wTInr9rb3HLIhxtR3Ih+5Gkah2gv" +
				"utTiwMQeBoVFsqE1qbcQ1+x+O5F/12fZEcm+q6lIvjv282uTQYYDnzpDuYqZIEdN" +
				"3gVjtJPtoC2wNvmpNZB4UV6nyFwRDa/Z3SOgiN5yrw5e1fkcC02GVuyD9sBtAJSB" +
				"burtW5OggU4iFMJ66bqdOobtlvrTBsD9tT8AvZW/wQKBgQDHbbJa95CEVZUPWcm9" +
				"JWTiczIhga5Y3ZM90HlZBOo83xzOF2Au15lCIl7Yn9MfzM4S1xoFpriNX+CquSFk" +
				"AAEQH0iHMqQjMYfbtmXhx0mqtjOz2usLwqjsd6XfQJHKL7XwDxW84TY61/vWwQr6" +
				"/bjOR15wegWFusf8LN2yut7XwwKBgQDV+/APWWvFEi+2e+WsE8oigHEy5awNWv8I" +
				"OTdOQ1Q7OuSkfv/Lxy4dga9lAUr7eDxhriYa18KLR3MebnkRDQVyrxinlY4qCCU7" +
				"SlY1K79OZPlfRu/S2kuA7K+8N9PkT/V7wvdgTI9mMtdJj9zE0qro9Ur9AGSkmzIR" +
				"5t1vvmt64QKBgAuUHd/UMcrNITtj7ieSLTpMj+OMIPA95ReYrAL0Gxlvpr98cfQm" +
				"RlqlnjYbiWl2PZywam1bkal7oJKo7vxcV7N07YQT952neYjTHTUvmeJUc8oEctMa" +
				"+S3JgJLmr9A6VujaJ1vxA3IFKjT8vkN2Sa2ITT5gh0ONZaEJhdGjsd57AoGBAM0G" +
				"VI/QZNLwxuh4s6l5WJ5QJKXYq04sltkBQT1qg2Uw22vFBz/vev7oh+4mG/rvzCLn" +
				"Yjkr64nZjrJktPkiWcr1e5DuWcVqAopZgln1rZnmY4zngdesMtW3cfXMI+jIt/O5" +
				"7Z3GHUuVgPNJtQScuQb2J8BbxRJ2ZLYEVry/XWnhAoGBALU3atePkLStC92M6Hca" +
				"VQIQZWXb03/ZRr8HrtX3Gn8Pc39umT2lYSwdq2IaP0iBBL40a9/if01MwZ6PiDKJ" +
				"GAK4T6m2DWB3F6fVTAIh9iYfDEAdnVbKO9PJqK8UIfWNHRXp0Q1jh/8MgGNXghv8" +
				"lEnUGeJucKII47X0DZBUDPT1\n" +
				"-----END RSA PRIVATE KEY-----",
			wantErr: errors.New("x509: failed to parse private key (use ParsePKCS8PrivateKey instead for this key format)"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, vErr := ParseRsaPrivateKey([]byte(tt.pemKey))
			assert.Equal(t, tt.wantErr, vErr)
		})
	}
}

func TestSetUpConfiguration(t *testing.T) {

	tests := []struct {
		name     string
		fileName string
		wantErr  error
	}{
		{
			name:     "trial run",
			fileName: "",
			wantErr:  nil,
		},
		{
			name:     "wrong format",
			fileName: "file",
			wantErr:  errors.New("config file should have .json extension"),
		},
		{
			name:     "missing file",
			fileName: "file.json",
			wantErr:  &fs.PathError{Op: "open", Path: "file.json", Err: syscall.ENOENT},
		},
	}
	for _, tt := range tests {
		serverConfig := config.NewServerConfig()
		t.Run(tt.name, func(t *testing.T) {
			_, vErr := SetUpConfiguration(&tt.fileName, serverConfig)
			assert.Equal(t, tt.wantErr, vErr)

		})
	}
}

func TestSetUpCryptoKey(t *testing.T) {

	tests := []struct {
		name      string
		cryptoKey string
		wantErr   error
	}{
		{
			name:      "trial run",
			cryptoKey: "",
			wantErr:   nil,
		},
		{
			name:      "wrong format",
			cryptoKey: "file",
			wantErr:   errors.New("private key file should have .pem extension"),
		},
		{
			name:      "missing file",
			cryptoKey: "file.pem",
			wantErr:   &fs.PathError{Op: "open", Path: "file.pem", Err: syscall.ENOENT},
		},
	}
	for _, tt := range tests {
		serverConfig := config.NewServerConfig()
		t.Run(tt.name, func(t *testing.T) {
			_, vErr := SetUpCryptoKey(&tt.cryptoKey, serverConfig)
			assert.Equal(t, tt.wantErr, vErr)

		})
	}
}

func TestSetUpDataStorage(t *testing.T) {
	tests := []struct {
		name         string
		storeFile    string
		restoreValue bool
	}{
		{
			name:         "no data upload",
			storeFile:    "",
			restoreValue: false,
		},
		{
			name:         "uploading from json",
			storeFile:    "/tmp/devops-metrics-db.json",
			restoreValue: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connStr := ""
			storeInt := 20
			storeParameter := "20"
			SetUpDataStorage(&connStr, &tt.storeFile, tt.restoreValue, storeInt, &storeParameter)

		})
	}
}

func TestFatalSetUpCryptoKey(t *testing.T) {

	cmd := exec.Command(os.Args[0], "-a address")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() && e.ExitCode() == 2 {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 2", err)
}

func ExampleInitializeRouter() {

	r := InitializeRouter(nil)
	ts := httptest.NewServer(r)
	defer ts.Close()
	floatValue := 2.0

	metricsObj := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &floatValue,
	}
	body, _ := json.Marshal(metricsObj)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/update/", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error happened in creating request. Err: %s", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error happened in posting request. Err: %s", err)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error happened in reading response body. Err: %s", err)
		return
	}
	defer resp.Body.Close()
	log.Print(respBody)

}

func testRequest(t *testing.T, ts *httptest.Server, path string, metricsObj metrics.Metrics) (*http.Response, string) {

	body, _ := json.Marshal(metricsObj)

	req, err := http.NewRequest(http.MethodPost, ts.URL+path, bytes.NewBuffer(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {

	r := mux.NewRouter()

	handlersWithKey := handlers.NewWrapperJSONStruct()

	r.HandleFunc("/update/", handlersWithKey.UpdateJSONHandler)
	r.HandleFunc("/value/", handlersWithKey.ValueJSONHandler)
	r.HandleFunc("/update/{type}/{name}/{value}", handlersWithKey.UpdateStringHandler)
	r.HandleFunc("/value/{type}/{name}", handlersWithKey.ValueStringHandler)

	r.HandleFunc("/", handlersWithKey.GenericHandler)
	r.Use(middleware.GzipHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()
	floatValue := 2.0

	metricsObj := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &floatValue,
	}

	resp, body := testRequest(t, ts, "/update/", metricsObj)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"status":"ok"}`, body)

	wrongMetrics := metrics.Metrics{
		ID:    "Alloc",
		MType: "othertype",
		Value: &floatValue,
	}

	resp, body = testRequest(t, ts, "/update/", wrongMetrics)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, `{"status":"invalid type"}`, body)

}

func TestInitializeRouter(t *testing.T) {

	tests := []struct {
		name string
	}{
		{
			name: "trial run",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitializeRouter(nil)

		})
	}
}

func TestShutdownGracefully(t *testing.T) {

	tests := []struct {
		name      string
		srv       *http.Server
		storeFile string
		connStr   string
	}{
		{
			name:      "trial run",
			storeFile: "/tmp/devops-metrics-db.json",
			connStr:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.srv = &http.Server{}
			ShutdownGracefully(tt.srv, storeFile, connStr)
		})
	}
}
