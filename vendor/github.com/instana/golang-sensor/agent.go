package instana

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	agentDiscoveryURL = "/com.instana.plugin.golang.discovery"
	agentTracesURL    = "/com.instana.plugin.golang/traces."
	agentDataURL      = "/com.instana.plugin.golang."
	agentEventURL     = "/com.instana.plugin.generic.event"
	agentDefaultHost  = "localhost"
	agentDefaultPort  = 42699
	agentHeader       = "Instana Agent"
)

type agentResponse struct {
	Pid    uint32 `json:"pid"`
	HostID string `json:"agentUuid"`
}

type discoveryS struct {
	PID   int      `json:"pid"`
	Name  string   `json:"name"`
	Args  []string `json:"args"`
	Fd    string   `json:"fd"`
	Inode string   `json:"inode"`
}

type fromS struct {
	PID    string `json:"e"`
	HostID string `json:"h"`
}

type agentS struct {
	sensor *sensorS
	fsm    *fsmS
	from   *fromS
	host   string
}

func (r *agentS) init() {
	r.fsm = r.initFsm()
	r.setFrom(&fromS{})
}

func (r *agentS) makeURL(prefix string) string {
	return r.makeHostURL(r.host, prefix)
}

func (r *agentS) makeHostURL(host string, prefix string) string {
	var port int
	if r.sensor.options.AgentPort == 0 {
		port = agentDefaultPort
	} else {
		port = r.sensor.options.AgentPort
	}

	return r.makeFullURL(host, port, prefix)
}

func (r *agentS) makeFullURL(host string, port int, prefix string) string {
	var buffer bytes.Buffer

	buffer.WriteString("http://")
	buffer.WriteString(host)
	buffer.WriteString(":")
	buffer.WriteString(strconv.Itoa(port))
	buffer.WriteString(prefix)
	if prefix[len(prefix)-1:] == "." && r.from.PID != "" {
		buffer.WriteString(r.from.PID)
	}

	return buffer.String()
}

func (r *agentS) head(url string) (string, error) {
	return r.request(url, "HEAD", nil)
}

func (r *agentS) request(url string, method string, data interface{}) (string, error) {
	return r.fullRequestResponse(url, method, data, nil, "")
}

func (r *agentS) requestResponse(url string, method string, data interface{}, ret interface{}) (string, error) {
	return r.fullRequestResponse(url, method, data, ret, "")
}

func (r *agentS) requestHeader(url string, method string, header string) (string, error) {
	return r.fullRequestResponse(url, method, nil, nil, header)
}

func (r *agentS) fullRequestResponse(url string, method string, data interface{}, body interface{}, header string) (string, error) {
	var j []byte
	var ret string
	var err error
	var resp *http.Response
	var req *http.Request
	if data != nil {
		j, err = json.Marshal(data)
	}

	if err == nil {
		if j != nil {
			req, err = http.NewRequest(method, url, bytes.NewBuffer(j))
		} else {
			req, err = http.NewRequest(method, url, nil)
		}

		if err == nil {
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err = client.Do(req)
			if err == nil {
				defer resp.Body.Close()

				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					err = errors.New(resp.Status)
				} else {

					log.debug("agent response:", url, resp.Status)

					if body != nil {
						var b []byte
						b, err = ioutil.ReadAll(resp.Body)
						json.Unmarshal(b, body)
					}

					if header != "" {
						ret = resp.Header.Get(header)
					}
				}
			}
		}
	}

	if err != nil {
		log.error(err)
	}

	return ret, err
}

func (r *agentS) setFrom(from *fromS) {
	r.from = from
}

func (r *agentS) setHost(host string) {
	r.host = host
}

func (r *agentS) reset() {
	r.fsm.reset()
}

func (r *sensorS) initAgent() *agentS {

	log.debug("initializing agent")

	ret := new(agentS)
	ret.sensor = r
	ret.init()

	return ret
}
