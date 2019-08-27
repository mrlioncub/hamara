package grafana

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

type RestClient struct {
	baseURL string
	key     string
	client  *http.Client
}

func NewRestClient(host string, key string) *RestClient {
	baseURL, _ := url.Parse(host)
	key = fmt.Sprintf("Bearer %s", key)
	return &RestClient{baseURL: baseURL.String(), key: key, client: http.DefaultClient}
}

func (r *RestClient) Get(query string, params url.Values) ([]byte, int, error) {
	return r.doRequest("GET", query, params, nil)
}

func (r *RestClient) doRequest(method, query string, params url.Values, buf io.Reader) ([]byte, int, error) {
	u, _ := url.Parse(r.baseURL)
	u.Path = path.Join(u.Path, query)
	if params != nil {
		u.RawQuery = params.Encode()
	}
	req, err := http.NewRequest(method, u.String(), buf)
	req.Header.Set("Authorization", r.key)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return data, resp.StatusCode, err
}

func (r *RestClient) GetAllDatasources() ([]*DataSource, error) {
	var (
		raw         []byte
		datasources []*DataSource
		code        int
		err         error
	)

	if raw, code, err = r.Get("api/datasources", nil); err != nil {
		return nil, err
	}

	if code != 200 {
		return nil, fmt.Errorf("HTTP error %d", code)
	}

	if err = json.Unmarshal(raw, &datasources); err != nil {
		return nil, err
	}

	for idx, ds := range datasources {
		var newDs DataSource
		if newDs, err = r.GetDatasource(ds.Id); err != nil {
			return nil, err
		}

		// assign the current datasource with the new datasource with json data field
		ds = &newDs
		datasources[idx] = ds

		ds.SecureJsonData = make(map[string]string)
		for key, value := range ds.SecureJsonFields {
			if value {
				ds.SecureJsonData[key] = fmt.Sprintf("$%s_%s", ds.Name, key)
			}
		}
	}

	return datasources, err
}

func (r *RestClient) GetDatasource(id int64) (DataSource, error) {
	var (
		raw  []byte
		ds   DataSource
		code int
		err  error
	)

	if raw, code, err = r.Get(fmt.Sprintf("api/datasources/%d", id), nil); err != nil {
		return ds, err
	}
	if code != 200 {
		return ds, fmt.Errorf("HTTP error %d", code)
	}

	json.Unmarshal(raw, &ds)
	return ds, err
}
