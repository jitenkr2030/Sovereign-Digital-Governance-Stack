package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
	"github.com/gorilla/mux"
	"github.com/labstack/echo/v4"
)

// HTTPServer defines the interface for HTTP server operations.
// This interface provides methods for serving HTTP requests with middleware support.
type HTTPServer interface {
	// Start starts the HTTP server
	Start(ctx context.Context, addr string) error

	// StartTLS starts the HTTP server with TLS
	StartTLS(ctx context.Context, addr string, certFile, keyFile string) error

	// Stop gracefully stops the HTTP server
	Stop(ctx context.Context) error

	// GetAddr returns the server address
	GetAddr() string

	// GetRouter returns the router
	GetRouter() interface{}

	// Use adds middleware
	Use(middleware ...interface{})

	// GET registers a GET handler
	GET(path string, handler interface{})

	// POST registers a POST handler
	POST(path string, handler interface{})

	// PUT registers a PUT handler
	PUT(path string, handler interface{})

	// PATCH registers a PATCH handler
	PATCH(path string, handler interface{})

	// DELETE registers a DELETE handler
	DELETE(path string, handler interface{})

	// OPTIONS registers an OPTIONS handler
	OPTIONS(path string, handler interface{})

	// HEAD registers a HEAD handler
	HEAD(path string, handler interface{})
}

// HTTPHandler defines the interface for HTTP request handlers.
type HTTPHandler interface {
	// Handle handles an HTTP request
	Handle(c echo.Context) error

	// Method returns the HTTP method
	Method() string

	// Path returns the handler path
	Path() string

	// Middleware returns handler middleware
	Middleware() []interface{}
}

// HTTPRequest defines the interface for HTTP request operations.
type HTTPRequest interface {
	// Context returns the request context
	Context() context.Context

	// Method returns the request method
	Method() string

	// Path returns the request path
	Path() string

	// QueryParams returns query parameters
	QueryParams() map[string][]string

	// QueryParam returns a query parameter
	QueryParam(name string) string

	// PathParams returns path parameters
	PathParams() map[string]string

	// PathParam returns a path parameter
	PathParam(name string) string

	// Header returns a header value
	Header(name string) string

	// Headers returns all headers
	Headers() map[string][]string

	// Body returns the request body
	Body() ([]byte, error)

	// Bind binds the request body to a struct
	Bind(i interface{}) error

	// Validate validates the request
	Validate() error

	// Get gets a context value
	Get(key string) interface{}

	// Set sets a context value
	Set(key string, value interface{})
}

// HTTPResponse defines the interface for HTTP response operations.
type HTTPResponse interface {
	// JSON sends a JSON response
	JSON(statusCode int, i interface{}) error

	// JSONBlob sends a JSON blob response
	JSONBlob(statusCode int, blob []byte) error

	// XML sends an XML response
	XML(statusCode int, i interface{}) error

	// HTML sends an HTML response
	HTML(statusCode int, html string) error

	// HTMLBlob sends an HTML blob response
	HTMLBlob(statusCode int, html []byte) error

	// String sends a string response
	String(statusCode int, format string, i ...interface{}) error

	// Stream sends a stream response
	Stream(statusCode int, contentType string, r io.Reader) error

	// File sends a file response
	.File(file string) error

	// Attachment sends an attachment response
	Attachment(file string, name string) error

	// Redirect redirects the request
	Redirect(statusCode int, url string) error

	// NoContent sends a no content response
	NoContent(statusCode int) error

	// SetHeader sets a response header
	SetHeader(name, value string)

	// SetContentType sets the content type
	SetContentType(contentType string)

	// Status sets the status code
	Status(statusCode int)

	// GetStatus gets the status code
	GetStatus() int
}

// HTTPMiddleware defines the interface for HTTP middleware.
type HTTPMiddleware interface {
	// Process processes the request
	Process(c echo.Context, next echo.HandlerFunc) error

	// Priority returns the middleware priority
	Priority() int

	// Name returns the middleware name
	Name() string
}

// HTTPAuthMiddleware defines authentication middleware operations.
type HTTPAuthMiddleware interface {
	// Authenticate authenticates the request
	Authenticate(c echo.Context) (context.Context, error)

	// Authorize authorizes the request
	Authorize(c echo.Context, resource string, action string) error

	// GetUser gets the authenticated user
	GetUser(c echo.Context) (*domain.User, error)

	// GetToken gets the token from the request
	GetToken(c echo.Context) (string, error)

	// UnaryInterceptor returns a middleware function
	UnaryInterceptor() echo.MiddlewareFunc
}

// HTTPRateLimiter defines rate limiting middleware operations.
type HTTPRateLimiter interface {
	// Limit limits the request
	Limit(c echo.Context) (bool, error)

	// GetLimit gets the rate limit for a key
	GetLimit(c echo.Context) (*domain.RateLimit, error)

	// SetLimit sets a rate limit
	SetLimit(key string, limit *domain.RateLimit)

	// UnaryInterceptor returns a middleware function
	UnaryInterceptor() echo.MiddlewareFunc

	// GetMetrics gets rate limiting metrics
	GetMetrics(ctx context.Context) (*domain.RateLimitMetrics, error)
}

// HTTPRequestLogger defines request logging middleware operations.
type HTTPRequestLogger interface {
	// Log logs the request
	Log(c echo.Context, rw echo.Response, d time.Duration)

	// GetFormatter returns the log formatter
	GetFormatter() HTTPLogFormatter

	// SetFormatter sets the log formatter
	SetFormatter(formatter HTTPLogFormatter)

	// UnaryInterceptor returns a middleware function
	UnaryInterceptor() echo.MiddlewareFunc
}

// HTTPLogFormatter defines log formatting operations.
type HTTPLogFormatter interface {
	// Format formats a log entry
	Format(c echo.Context, rw echo.Response, d time.Duration, err error) map[string]interface{}

	// FormatError formats an error
	FormatError(err error) map[string]interface{}
}

// HTTPCORS defines CORS middleware operations.
type HTTPCORS interface {
	// AllowOrigin returns the allowed origin
	AllowOrigin(c echo.Context) string

	// AllowMethods returns the allowed methods
	AllowMethods() []string

	// AllowHeaders returns the allowed headers
	AllowHeaders() []string

	// ExposeHeaders returns the exposed headers
	ExposeHeaders() []string

	// MaxAge returns the max age
	MaxAge() int

	// AllowCredentials returns whether credentials are allowed
	AllowCredentials() bool

	// UnaryInterceptor returns a middleware function
	UnaryInterceptor() echo.MiddlewareFunc
}

// HTTPRecovery defines recovery middleware operations.
type HTTPRecovery interface {
	// Recover recovers from a panic
	Recover(c echo.Context, err error) error

	// GetRecoveryHandler returns the recovery handler
	GetRecoveryHandler() func(c echo.Context, err error) error

	// SetRecoveryHandler sets the recovery handler
	SetRecoveryHandler(handler func(c echo.Context, err error) error)

	// UnaryInterceptor returns a middleware function
	UnaryInterceptor() echo.MiddlewareFunc
}

// HTTPSecurity defines security middleware operations.
type HTTPSecurity interface {
	// AddSecurityHeaders adds security headers
	AddSecurityHeaders(c echo.Context)

	// GetCSP returns the content security policy
	GetCSP() string

	// SetCSP sets the content security policy
	SetCSP(csp string)

	// UnaryInterceptor returns a middleware function
	UnaryInterceptor() echo.MiddlewareFunc
}

// HTTPRouter defines routing operations.
type HTTPRouter interface {
	// GET registers a GET route
	GET(path string, handler interface{})

	// POST registers a POST route
	POST(path string, handler interface{})

	// PUT registers a PUT route
	PUT(path string, handler interface{})

	// PATCH registers a PATCH route
	PATCH(path string, handler interface{})

	// DELETE registers a DELETE route
	DELETE(path string, handler interface{})

	// OPTIONS registers an OPTIONS route
	OPTIONS(path string, handler interface{})

	// HEAD registers a HEAD route
	HEAD(path string, handler interface{})

	// Any registers a route for all methods
	Any(path string, handler interface{})

	// Group creates a route group
	Group(prefix string) HTTPRouter

	// Use registers middleware for the router
	Use(middleware ...interface{})

	// Static registers a static file route
	Static(path, dir string)

	// File registers a file route
	 File(path, file string)

	// Route registers a route with a name
	Route(path, name string, method string, handler interface{}) HTTPRoute

	// Routes returns all registered routes
	Routes() []HTTPRoute
}

// HTTPRoute defines route operations.
type HTTPRoute interface {
	// Name sets the route name
	Name(name string) HTTPRoute

	// Path returns the route path
	Path() string

	// Method returns the route method
	Method() string

	// Handler returns the route handler
	Handler() interface{}

	// Middleware returns the route middleware
	Middleware() []interface{}

	// Summary sets the route summary
	Summary(summary string) HTTPRoute

	// Description sets the route description
	Description(description string) HTTPRoute

	// Tags sets the route tags
	Tags(tags ...string) HTTPRoute

	// OperationID sets the operation ID
	OperationID(operationID string) HTTPRoute

	// Deprecated marks the route as deprecated
	Deprecated() HTTPRoute

	// Validate validates the route
	Validate() error
}

// HTTPValidator defines request validation operations.
type HTTPValidator interface {
	// Validate validates a request
	Validate(c echo.Context, schema interface{}) error

	// ValidatePath validates path parameters
	ValidatePath(c echo.Context, schema interface{}) error

	// ValidateQuery validates query parameters
	ValidateQuery(c echo.Context, schema interface{}) error

	// ValidateHeader validates headers
	ValidateHeader(c echo.Context, schema interface{}) error

	// ValidateBody validates the request body
	ValidateBody(c echo.Context, schema interface{}) error

	// GetErrors returns validation errors
	GetErrors() []*domain.ValidationError
}

// HTTPErrorHandler defines custom error handling operations.
type HTTPErrorHandler interface {
	// HandleError handles an error
	HandleError(c echo.Context, err error) error

	// GetErrorResponse gets the error response
	GetErrorResponse(err error) *domain.ErrorResponse

	// GetHTTPError gets the HTTP error
	GetHTTPError(err error) *domain.HTTPError

	// SetHTTPError sets the HTTP error
	SetHTTPError(err *domain.HTTPError)
}

// HTTPClient defines HTTP client operations.
type HTTPClient interface {
	// Get performs a GET request
	Get(ctx context.Context, url string, opts ...HTTPOption) (*HTTPResponse, error)

	// Post performs a POST request
	Post(ctx context.Context, url string, body interface{}, opts ...HTTPOption) (*HTTPResponse, error)

	// Put performs a PUT request
	Put(ctx context.Context, url string, body interface{}, opts ...HTTPOption) (*HTTPResponse, error)

	// Patch performs a PATCH request
	Patch(ctx context.Context, url string, body interface{}, opts ...HTTPOption) (*HTTPResponse, error)

	// Delete performs a DELETE request
	Delete(ctx context.Context, url string, opts ...HTTPOption) (*HTTPResponse, error)

	// Head performs a HEAD request
	Head(ctx context.Context, url string, opts ...HTTPOption) (*HTTPResponse, error)

	// Do performs a custom request
	Do(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)

	// SetBaseURL sets the base URL
	SetBaseURL(url string)

	// SetTimeout sets the timeout
	SetTimeout(timeout time.Duration)

	// SetMiddleware sets middleware
	SetMiddleware(middleware ...HTTPMiddleware)

	// GetMetrics gets client metrics
	GetMetrics(ctx context.Context) (*domain.HTTPClientMetrics, error)
}

// HTTPOption defines HTTP client options.
type HTTPOption interface {
	Apply(opts *domain.HTTPOptions)
}

// HTTPRequestBuilder defines request building operations.
type HTTPRequestBuilder interface {
	// Method sets the request method
	Method(method string) HTTPRequestBuilder

	// URL sets the request URL
	URL(url string) HTTPRequestBuilder

	// Body sets the request body
	Body(body interface{}) HTTPRequestBuilder

	// QueryParam adds a query parameter
	QueryParam(name, value string) HTTPRequestBuilder

	// QueryParams adds query parameters
	QueryParams(params map[string]string) HTTPRequestBuilder

	// Header adds a header
	Header(name, value string) HTTPRequestBuilder

	// Headers adds headers
	Headers(headers map[string]string) HTTPRequestBuilder

	// ContentType sets the content type
	ContentType(contentType string) HTTPRequestBuilder

	// Accept sets the accept header
	Accept(accept string) HTTPRequestBuilder

	// Authorization sets the authorization header
	Authorization(auth string) HTTPRequestBuilder

	// BearerToken sets the bearer token
	BearerToken(token string) HTTPRequestBuilder

	// BasicAuth sets basic authentication
	BasicAuth(username, password string) HTTPRequestBuilder

	// Timeout sets the timeout
	Timeout(timeout time.Duration) HTTPRequestBuilder

	// Build builds the request
	Build() HTTPRequest

	// Do executes the request
	Do(ctx context.Context) (*HTTPResponse, error)
}

// HTTPSwagger defines Swagger/OpenAPI operations.
type HTTPSwagger interface {
	// GetSpec returns the OpenAPI spec
	GetSpec() []byte

	// Register registers the API with Swagger
	Register(title, version string) error

	// AddRoute adds a route to Swagger
	AddRoute(route HTTPRoute, schema interface{}) error

	// AddDefinition adds a definition
	AddDefinition(name string, definition interface{}) error

	// AddSecurityScheme adds a security scheme
	AddSecurityScheme(name string, scheme *domain.SecurityScheme) error
}

// HTTPMetrics defines HTTP metrics operations.
type HTTPMetrics interface {
	// IncRequests increments the request counter
	IncRequests(method, path, status string)

	// ObserveLatency observes the request latency
	ObserveLatency(method, path string, latency time.Duration)

	// ObserveRequestSize observes the request size
	ObserveRequestSize(method, path string, size int64)

	// ObserveResponseSize observes the response size
	ObserveResponseSize(method, path string, size int64)

	// IncErrors increments the error counter
	IncErrors(method, path string, errorType string)

	// GetMetrics gets metrics
	GetMetrics() *domain.HTTPMetrics

	// Reset resets metrics
	Reset()
}

// HTTPHealthCheck defines health check endpoint operations.
type HTTPHealthCheck interface {
	// HealthCheck performs a health check
	HealthCheck(c echo.Context) error

	// ReadyCheck performs a readiness check
	ReadyCheck(c echo.Context) error

	// LiveCheck performs a liveness check
	LiveCheck(c echo.Context) error

	// RegisterChecker registers a health checker
	RegisterChecker(name string, checker interface{})

	// GetHealthStatus gets the health status
	GetHealthStatus() *domain.HealthStatus
}

// PolicyHTTPHandler defines HTTP handler operations for policies.
type PolicyHTTPHandler interface {
	// CreatePolicy handles policy creation
	CreatePolicy(c echo.Context) error

	// GetPolicy handles policy retrieval
	GetPolicy(c echo.Context) error

	// ListPolicies handles policy listing
	ListPolicies(c echo.Context) error

	// UpdatePolicy handles policy update
	UpdatePolicy(c echo.Context) error

	// DeletePolicy handles policy deletion
	DeletePolicy(c echo.Context) error

	// SearchPolicies handles policy search
	SearchPolicies(c echo.Context) error

	// ValidatePolicy handles policy validation
	ValidatePolicy(c echo.Context) error
}

// EnforcementHTTPHandler defines HTTP handler operations for enforcements.
type EnforcementHTTPHandler interface {
	// CreateEnforcement handles enforcement creation
	CreateEnforcement(c echo.Context) error

	// GetEnforcement handles enforcement retrieval
	GetEnforcement(c echo.Context) error

	// ListEnforcements handles enforcement listing
	ListEnforcements(c echo.Context) error

	// UpdateEnforcement handles enforcement update
	UpdateEnforcement(c echo.Context) error

	// TransitionEnforcement handles enforcement transition
	TransitionEnforcement(c echo.Context) error

	// AssignEnforcement handles enforcement assignment
	AssignEnforcement(c echo.Context) error
}

// StateHTTPHandler defines HTTP handler operations for states.
type StateHTTPHandler interface {
	// GetState handles state retrieval
	GetState(c echo.Context) error

	// GetStateHistory handles state history retrieval
	GetStateHistory(c echo.Context) error

	// TransitionState handles state transition
	TransitionState(c echo.Context) error

	// ListStates handles state listing
	ListStates(c echo.Context) error

	// ValidateTransition handles transition validation
	ValidateTransition(c echo.Context) error
}

// InterventionHTTPHandler defines HTTP handler operations for interventions.
type InterventionHTTPHandler interface {
	// CreateIntervention handles intervention creation
	CreateIntervention(c echo.Context) error

	// GetIntervention handles intervention retrieval
	GetIntervention(c echo.Context) error

	// ListInterventions handles intervention listing
	ListInterventions(c echo.Context) error

	// UpdateIntervention handles intervention update
	UpdateIntervention(c echo.Context) error

	// TransitionIntervention handles intervention transition
	TransitionIntervention(c echo.Context) error

	// ExecuteIntervention handles intervention execution
	ExecuteIntervention(c echo.Context) error
}

// ComplianceHTTPHandler defines HTTP handler operations for compliance.
type ComplianceHTTPHandler interface {
	// CheckCompliance handles compliance check
	CheckCompliance(c echo.Context) error

	// GetComplianceStatus handles compliance status retrieval
	GetComplianceStatus(c echo.Context) error

	// ListComplianceRecords handles compliance record listing
	ListComplianceRecords(c echo.Context) error

	// ReportViolation handles violation reporting
	ReportViolation(c echo.Context) error

	// RequestExemption handles exemption request
	RequestExemption(c echo.Context) error

	// GetComplianceMetrics handles compliance metrics retrieval
	GetComplianceMetrics(c echo.Context) error

	// GenerateReport handles compliance report generation
	GenerateReport(c echo.Context) error
}
