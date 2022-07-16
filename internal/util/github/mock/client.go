package mock

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// EndpointPattern models the GitHub's API endpoints
type EndpointPattern struct {
	Pattern string // eg. "/repos/{owner}/{repo}/actions/artifacts"
	Method  string // "GET", "POST", "PATCH", etc
}

// MockBackendOption is used to configure the *mux.router
// for the mocked backend
type MockBackendOption func(*mux.Router)

type FIFOReponseHandler struct {
	Responses    [][]byte
	CurrentIndex int
}

// ServeHTTP implementation of `http.Handler`
func (srh *FIFOReponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if srh.CurrentIndex > len(srh.Responses) {
		panic(fmt.Sprintf(
			"go-github-mock: no more mocks available for %s",
			r.URL.Path,
		))
	}

	defer func() {
		srh.CurrentIndex++
	}()

	w.Write(srh.Responses[srh.CurrentIndex])
}

type PaginatedReponseHandler struct {
	ResponsePages [][]byte
}

func (prh *PaginatedReponseHandler) getCurrentPage(r *http.Request) int {
	strPage := r.URL.Query().Get("page")

	if strPage == "" {
		return 1
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))

	if err == nil {
		return page
	}

	// this should never happen as the request is being made by the SDK
	panic(fmt.Sprintf("invalid page: %s", strPage))
}

func (prh *PaginatedReponseHandler) generateLinkHeader(
	w http.ResponseWriter,
	r *http.Request,
) {
	currentPage := prh.getCurrentPage(r)
	lastPage := len(prh.ResponsePages)

	buf := bytes.NewBufferString(`<?page=1>; rel="first",`)
	buf.WriteString(fmt.Sprintf(`<?page=%d>; rel="last",`, lastPage))

	if currentPage < lastPage {
		// when resp.NextPage == 0, it means no more pages
		// which is basically as not setting it in the response
		buf.WriteString(fmt.Sprintf(`<?page=%d>; rel="next",`, currentPage+1))
	}

	if currentPage > 1 {
		buf.WriteString(fmt.Sprintf(`<?page=%d>; rel="prev",`, currentPage-1))
	}

	w.Header().Add("Link", buf.String())
}

// ServeHTTP implementation of `http.Handler`
func (prh *PaginatedReponseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prh.generateLinkHeader(w, r)
	w.Write(prh.ResponsePages[prh.getCurrentPage(r)-1])
}

// EnforceHostRoundTripper rewrites all requests with the given `Host`.
type EnforceHostRoundTripper struct {
	Host                 string
	UpstreamRoundTripper http.RoundTripper
}

// RoundTrip implementation of `http.RoundTripper`
func (efrt *EnforceHostRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	splitHost := strings.Split(efrt.Host, "://")
	r.URL.Scheme = splitHost[0]
	r.URL.Host = splitHost[1]

	return efrt.UpstreamRoundTripper.RoundTrip(r)
}

func NewMockedHTTPClient(options ...MockBackendOption) *http.Client {
	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteError(
			w,
			http.StatusNotFound,
			fmt.Sprintf("mock response not found for %s", r.URL.Path),
		)
	})

	for _, o := range options {
		o(router)
	}

	mockServer := httptest.NewServer(router)

	c := mockServer.Client()

	c.Transport = &EnforceHostRoundTripper{
		Host:                 mockServer.URL,
		UpstreamRoundTripper: mockServer.Client().Transport,
	}

	return c
}

func WithRequestMatch(
	ep EndpointPattern,
	responsesFIFO ...interface{},
) MockBackendOption {
	var responses [][]byte

	for _, r := range responsesFIFO {
		responses = append(responses, MustMarshal(r))
	}

	return WithRequestMatchHandler(ep, &FIFOReponseHandler{
		Responses: responses,
	})
}

func WithRequestMatchHandler(
	ep EndpointPattern,
	handler http.Handler,
) MockBackendOption {
	return func(router *mux.Router) {
		router.Handle(ep.Pattern, handler).Methods(ep.Method)
	}
}
