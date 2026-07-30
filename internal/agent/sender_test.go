package agent

import (
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func Ptr[T any](v T) *T {
	return &v
}

func TestSendMetric(t *testing.T) {
	type want struct {
		result    string
		errNotNil bool
	}
	type response struct {
		status int
		body   string
		err    error
	}
	tests := []struct {
		name     string
		metric   models.Metrics
		response response
		want     want
	}{
		{
			name: "send runtime metric",
			metric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: Ptr(100.20),
			},
			response: response{
				status: http.StatusOK,
				body:   "100.20",
				err:    nil,
			},
			want: want{
				result:    "100.20",
				errNotNil: false,
			},
		},
		{
			name: "send runtime metric incorrect status",
			metric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: Ptr(100.20),
			},
			response: response{
				status: http.StatusNotFound,
				err:    nil,
			},
			want: want{
				result:    "",
				errNotNil: true,
			},
		},
	}
	config := GetDefaultAgentConfig()
	sender := NewSender(&config)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				headers := rw.Header()
				headers.Add("Content-Type", "text/plain")
				rw.WriteHeader(test.response.status)
				if test.response.body != "" {
					rw.Write([]byte(test.response.body))
				}
			}))
			defer server.Close()

			client := resty.New()
			sender.Config.ServerAddress = server.URL

			result, err := sender.SendMetric(client, &test.metric)
			assert.Equal(t, test.want.result, result)
			if test.want.errNotNil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSendMetricJSON(t *testing.T) {
	type want struct {
		result    string
		errNotNil bool
	}
	type response struct {
		status int
		body   string
		err    error
	}
	tests := []struct {
		name     string
		metric   models.Metrics
		response response
		want     want
	}{
		{
			name: "send runtime metric",
			metric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: Ptr(100.20),
			},
			response: response{
				status: http.StatusOK,
				body:   `{"id":"Alloc","type":"gauge","value":100.20}`,
				err:    nil,
			},
			want: want{
				result:    "100.2",
				errNotNil: false,
			},
		},
		{
			name: "send runtime metric incorrect status",
			metric: models.Metrics{
				ID:    "Alloc",
				MType: models.Gauge,
				Value: Ptr(100.20),
			},
			response: response{
				status: http.StatusNotFound,
				err:    nil,
			},
			want: want{
				result:    "",
				errNotNil: true,
			},
		},
	}

	config := GetDefaultAgentConfig()
	sender := NewSender(&config)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				headers := rw.Header()
				headers.Add("Content-Type", "application/json")
				headers.Set("Content-Encoding", "gzip")

				rw.WriteHeader(test.response.status)
				if test.response.body != "" {
					gz := gzip.NewWriter(rw)
					defer gz.Close()
					gz.Write([]byte(test.response.body))
				}
			}))
			defer server.Close()

			client := resty.New()
			sender.Config.ServerAddress = server.URL

			result, err := sender.SendMetricJSON(client, &test.metric)
			assert.Equal(t, test.want.result, result)
			if test.want.errNotNil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSendMetricsJSON(t *testing.T) {
	type want struct {
		result      string
		errNotNil   bool
		requestBody string
	}
	type response struct {
		status int
		body   string
		err    error
	}
	tests := []struct {
		name     string
		metrics  []models.Metrics
		response response
		want     want
	}{
		{
			name: "correct send metrics",
			metrics: []models.Metrics{
				{
					ID:    "TestGauge",
					MType: models.Gauge,
					Value: Ptr(100.20),
				},
				{
					ID:    "TestCounter",
					MType: models.Counter,
					Delta: Ptr(int64(10)),
				},
			},
			response: response{
				status: http.StatusOK,
				body:   `{"status":"ok","message":"successfully saved 2 metrics"}`,
				err:    nil,
			},
			want: want{
				result:      `successfully saved 2 metrics`,
				errNotNil:   false,
				requestBody: `[{"id":"TestGauge","type":"gauge","value":100.2},{"id":"TestCounter","type":"counter","delta":10}]`,
			},
		},
	}

	config := GetDefaultAgentConfig()
	sender := NewSender(&config)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// generate keys to encrypt and decrypt request
			privKeyFile, err := os.CreateTemp("", "priv-key-*.pem")
			if err != nil {
				panic(err)
			}
			defer os.Remove(privKeyFile.Name())
			pubKeyFile, err := os.CreateTemp("", "pub-key-*.pem")
			if err != nil {
				panic(err)
			}
			defer os.Remove(pubKeyFile.Name())

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				panic(err)
			}
			err = savePrivateKeyToFile(privKeyFile, privateKey)
			if err != nil {
				panic(err)
			}
			err = savePublicKeyToFile(pubKeyFile, &privateKey.PublicKey)
			if err != nil {
				panic(err)
			}

			// create sever simulation
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				headers := rw.Header()
				headers.Add("Content-Type", "application/json")
				headers.Set("Content-Encoding", "gzip")

				var requestBody []byte
				if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
					var gzReader *gzip.Reader
					gzReader, err = gzip.NewReader(req.Body)
					if err != nil {
						panic(err)
					}
					defer gzReader.Close()
					requestBody, err = io.ReadAll(gzReader)
					if err != nil {
						rw.WriteHeader(http.StatusBadRequest)
						return
					}

					requestBody, err = decryptBytes(privateKey, requestBody)
					if err != nil {
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
				// check that body same as expected
				assert.Equal(t, test.want.requestBody, string(requestBody))

				rw.WriteHeader(test.response.status)
				if test.response.body != "" {
					gz := gzip.NewWriter(rw)
					defer gz.Close()
					gz.Write([]byte(test.response.body))
				}
			}))
			defer server.Close()

			// setup client
			client := resty.New()
			sender.Config.ServerAddress = server.URL
			sender.Config.CryptoKey = pubKeyFile.Name()
			sender.loadPublicKeyIfExists()

			// send request and check response
			result, err := sender.SendMetricsJSON(client, test.metrics)
			assert.Equal(t, test.want.result, result)
			if test.want.errNotNil {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			// close temp files
			privKeyFile.Close()
			pubKeyFile.Close()
		})
	}
}

func savePrivateKeyToFile(file *os.File, privateKey *rsa.PrivateKey) error {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	return pem.Encode(file, privateKeyPEM)
}

func savePublicKeyToFile(file *os.File, publicKey *rsa.PublicKey) error {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	return pem.Encode(file, publicKeyPEM)
}

func decryptBytes(privateKey *rsa.PrivateKey, encryptedMessage []byte) ([]byte, error) {
	chunkSize := privateKey.Size()

	// If the encrypted message is the same size as a single chunk, decrypt directly
	if len(encryptedMessage) <= chunkSize {
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedMessage, nil)
	}

	// Split the encrypted message into chunks
	var decryptedData []byte
	for start := 0; start < len(encryptedMessage); start += chunkSize {
		end := start + chunkSize
		if end > len(encryptedMessage) {
			// If the last chunk is smaller than chunkSize, it might be a direct encryption
			// or we have incomplete data
			decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedMessage[start:], nil)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt final chunk: %w", err)
			}
			decryptedData = append(decryptedData, decryptedChunk...)
			break
		}

		chunk := encryptedMessage[start:end]
		decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt chunk: %w", err)
		}

		decryptedData = append(decryptedData, decryptedChunk...)
	}

	return decryptedData, nil
}
