package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// GRPCServer defines the interface for gRPC server operations.
// This interface provides methods for serving gRPC requests.
type GRPCServer interface {
	// Serve starts serving gRPC requests
	Serve(ctx context.Context) error

	// Stop gracefully stops the gRPC server
	Stop(ctx context.Context) error

	// ForceStop forces the gRPC server to stop
	ForceStop()

	// GetPort returns the listening port
	GetPort() int

	// RegisterService registers a gRPC service
	RegisterService(desc *grpc.ServiceDesc, impl interface{})
}

// GRPCClient defines the interface for gRPC client operations.
// This interface provides methods for making gRPC calls.
type GRPCClient interface {
	// Connect connects to the gRPC server
	Connect(ctx context.Context, address string, opts ...grpc.DialOption) error

	// Disconnect disconnects from the gRPC server
	Disconnect(ctx context.Context) error

	// IsConnected checks if the client is connected
	IsConnected() bool

	// GetConn returns the underlying gRPC connection
	GetConn() *grpc.ClientConn

	// Invoke performs a unary RPC call
	Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error

	// NewStream creates a streaming RPC call
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStreamInterface, error)

	// Close closes the client connection
	Close() error
}

// GRPCHealthServer defines the interface for gRPC health checking.
// This interface follows the standard gRPC health checking protocol.
type GRPCHealthServer interface {
	// Check performs a health check
	Check(ctx context.Context, service string) (*healthCheckResponse, error)

	// Watch watches for health status changes
	Watch(*healthCheckRequest, grpc.ServerStreamingServer[healthCheckResponse]) error

	// SetServingStatus sets the serving status for a service
	SetServingStatus(service string, servingStatus healthServingStatus)

	// SetNotServingStatus sets the not serving status for a service
	SetNotServingStatus(service string)

	// EnterShuttingDown enters shutting down state
	EnterShuttingDown()

	// Resets all serving status to NOT_SERVING
	Reset()
}

// GRPCAuthInterceptor defines the interface for gRPC authentication interceptors.
type GRPCAuthInterceptor interface {
	// Authenticate authenticates a gRPC request
	Authenticate(ctx context.Context, method string, md metadata.MD) (context.Context, error)

	// Authorize authorizes a gRPC request
	Authorize(ctx context.Context, method string, ctx context.Context) error

	// UnaryInterceptor returns a unary interceptor
	UnaryInterceptor() grpc.UnaryServerInterceptor

	// StreamInterceptor returns a stream interceptor
	StreamInterceptor() grpc.StreamServerInterceptor
}

// GRPCRateLimiter defines the interface for gRPC rate limiting.
type GRPCRateLimiter interface {
	// Allow checks if the request is allowed
	Allow(ctx context.Context, method string, p peer.Peer) bool

	// UnaryInterceptor returns a unary interceptor for rate limiting
	UnaryInterceptor() grpc.UnaryServerInterceptor

	// StreamInterceptor returns a stream interceptor for rate limiting
	StreamInterceptor() grpc.StreamServerInterceptor

	// GetMetrics returns rate limiting metrics
	GetMetrics(ctx context.Context) (*domain.RateLimitMetrics, error)

	// UpdateLimit updates the rate limit for a method
	UpdateLimit(ctx context.Context, method string, limit *domain.RateLimitConfig) error
}

// GRPCLogger defines the interface for gRPC logging.
type GRPCLogger interface {
	// LogRequest logs a gRPC request
	LogRequest(ctx context.Context, method string, req interface{}, duration time.Duration, code codes.Code)

	// LogResponse logs a gRPC response
	LogResponse(ctx context.Context, method string, resp interface{}, duration time.Duration, code codes.Code)

	// LogError logs a gRPC error
	LogError(ctx context.Context, method string, err error, code codes.Code)

	// UnaryInterceptor returns a unary interceptor for logging
	UnaryInterceptor() grpc.UnaryServerInterceptor

	// StreamInterceptor returns a stream interceptor for logging
	StreamInterceptor() grpc.StreamServerInterceptor
}

//GRPCCircuitBreaker defines the interface for gRPC circuit breaking.
type GRPCCircuitBreaker interface {
	// Allow checks if the request is allowed
	Allow(ctx context.Context, method string) bool

	// OnSuccess records a successful call
	OnSuccess(ctx context.Context, method string)

	// OnError records an error
	OnError(ctx context.Context, method string, err error)

	// State returns the circuit breaker state
	State(ctx context.Context, method string) domain.CircuitState

	// Reset resets the circuit breaker
	Reset(ctx context.Context, method string)

	// UnaryInterceptor returns a unary interceptor for circuit breaking
	UnaryInterceptor() grpc.UnaryServerInterceptor

	// StreamInterceptor returns a stream interceptor for circuit breaking
	StreamInterceptor() grpc.StreamServerInterceptor
}

// GRPCRecovery defines the interface for gRPC panic recovery.
type GRPCRecovery interface {
	// Recover handles panics in gRPC handlers
	Recover(ctx interface{}, p interface{}) (err error)

	// UnaryInterceptor returns a unary interceptor for recovery
	UnaryInterceptor() grpc.UnaryServerInterceptor

	// StreamInterceptor returns a stream interceptor for recovery
	StreamInterceptor() grpc.StreamServerInterceptor
}

// GRPCMetrics defines the interface for gRPC metrics collection.
type GRPCMetrics interface {
	// IncRequests increments the request counter
	IncRequests(ctx context.Context, method string, code codes.Code)

	// IncActiveRequests increments the active requests counter
	IncActiveRequests(ctx context.Context, method string)

	// DecActiveRequests decrements the active requests counter
	DecActiveRequests(ctx context.Context, method string)

	// ObserveLatency observes the request latency
	ObserveLatency(ctx context.Context, method string, latency time.Duration)

	// ObserveMessageSent observes a sent message
	ObserveMessageSent(ctx context.Context, method string, count int64)

	// ObserveMessageReceived observes a received message
	ObserveMessageReceived(ctx context.Context, method string, count int64)

	// UnaryInterceptor returns a unary interceptor for metrics
	UnaryInterceptor() grpc.UnaryServerInterceptor

	// StreamInterceptor returns a stream interceptor for metrics
	StreamInterceptor() grpc.StreamServerInterceptor
}

// GRPCAuthService defines the interface for authentication service operations.
type GRPCAuthService interface {
	// Login authenticates a user and returns a token
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)

	// Logout logs out a user
	Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error)

	// Refresh refreshes an access token
	Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResponse, error)

	// Validate validates a token
	Validate(ctx context.Context, req *ValidateRequest) (*ValidateResponse, error)

	// Register registers a new user
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)

	// GetUser gets user information
	GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error)

	// ChangePassword changes the user's password
	ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error)

	// GetRoles gets all roles
	GetRoles(ctx context.Context, req *GetRolesRequest) (*GetRolesResponse, error)

	// AssignRole assigns a role to a user
	AssignRole(ctx context.Context, req *AssignRoleRequest) (*AssignRoleResponse, error)

	// RemoveRole removes a role from a user
	RemoveRole(ctx context.Context, req *RemoveRoleRequest) (*RemoveRoleResponse, error)

	// GetPermissions gets all permissions
	GetPermissions(ctx context.Context, req *GetPermissionsRequest) (*GetPermissionsResponse, error)
}

// GRPCPolicyService defines the interface for policy service operations.
type GRPCPolicyService interface {
	// CreatePolicy creates a new policy
	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*CreatePolicyResponse, error)

	// GetPolicy gets a policy by ID
	GetPolicy(ctx context.Context, req *GetPolicyRequest) (*GetPolicyResponse, error)

	// ListPolicies lists all policies
	ListPolicies(ctx context.Context, req *ListPoliciesRequest) (*ListPoliciesResponse, error)

	// UpdatePolicy updates a policy
	UpdatePolicy(ctx context.Context, req *UpdatePolicyRequest) (*UpdatePolicyResponse, error)

	// DeletePolicy deletes a policy
	DeletePolicy(ctx context.Context, req *DeletePolicyRequest) (*DeletePolicyResponse, error)

	// SearchPolicies searches policies
	SearchPolicies(ctx context.Context, req *SearchPoliciesRequest) (*SearchPoliciesResponse, error)

	// ValidatePolicy validates a policy
	ValidatePolicy(ctx context.Context, req *ValidatePolicyRequest) (*ValidatePolicyResponse, error)
}

// GRPCEnforcementService defines the interface for enforcement service operations.
type GRPCEnforcementService interface {
	// CreateEnforcement creates a new enforcement action
	CreateEnforcement(ctx context.Context, req *CreateEnforcementRequest) (*CreateEnforcementResponse, error)

	// GetEnforcement gets an enforcement by ID
	GetEnforcement(ctx context.Context, req *GetEnforcementRequest) (*GetEnforcementResponse, error)

	// ListEnforcements lists all enforcements
	ListEnforcements(ctx context.Context, req *ListEnforcementsRequest) (*ListEnforcementsResponse, error)

	// UpdateEnforcement updates an enforcement
	UpdateEnforcement(ctx context.Context, req *UpdateEnforcementRequest) (*UpdateEnforcementResponse, error)

	// TransitionEnforcement transitions an enforcement to a new state
	TransitionEnforcement(ctx context.Context, req *TransitionEnforcementRequest) (*TransitionEnforcementResponse, error)

	// AssignEnforcement assigns an enforcement to a user
	AssignEnforcement(ctx context.Context, req *AssignEnforcementRequest) (*AssignEnforcementResponse, error)
}

// GRPCStateService defines the interface for state service operations.
type GRPCStateService interface {
	// GetState gets the current state of an entity
	GetState(ctx context.Context, req *GetStateRequest) (*GetStateResponse, error)

	// GetStateHistory gets the state history of an entity
	GetStateHistory(ctx context.Context, req *GetStateHistoryRequest) (*GetStateHistoryResponse, error)

	// TransitionState transitions an entity to a new state
	TransitionState(ctx context.Context, req *TransitionStateRequest) (*TransitionStateResponse, error)

	// ListStates lists all states
	ListStates(ctx context.Context, req *ListStatesRequest) (*ListStatesResponse, error)

	// ValidateTransition validates a state transition
	ValidateTransition(ctx context.Context, req *ValidateTransitionRequest) (*ValidateTransitionResponse, error)
}

// GRPCInterventionService defines the interface for intervention service operations.
type GRPCInterventionService interface {
	// CreateIntervention creates a new intervention
	CreateIntervention(ctx context.Context, req *CreateInterventionRequest) (*CreateInterventionResponse, error)

	// GetIntervention gets an intervention by ID
	GetIntervention(ctx context.Context, req *GetInterventionRequest) (*GetInterventionResponse, error)

	// ListInterventions lists all interventions
	ListInterventions(ctx context.Context, req *ListInterventionsRequest) (*ListInterventionsResponse, error)

	// UpdateIntervention updates an intervention
	UpdateIntervention(ctx context.Context, req *UpdateInterventionRequest) (*UpdateInterventionResponse, error)

	// TransitionIntervention transitions an intervention to a new state
	TransitionIntervention(ctx context.Context, req *TransitionInterventionRequest) (*TransitionInterventionResponse, error)

	// ExecuteIntervention executes an intervention action
	ExecuteIntervention(ctx context.Context, req *ExecuteInterventionRequest) (*ExecuteInterventionResponse, error)
}

// GRPCComplianceService defines the interface for compliance service operations.
type GRPCComplianceService interface {
	// CheckCompliance checks compliance for an entity
	CheckCompliance(ctx context.Context, req *CheckComplianceRequest) (*CheckComplianceResponse, error)

	// GetComplianceStatus gets the compliance status of an entity
	GetComplianceStatus(ctx context.Context, req *GetComplianceStatusRequest) (*GetComplianceStatusResponse, error)

	// ListComplianceRecords lists compliance records
	ListComplianceRecords(ctx context.Context, req *ListComplianceRecordsRequest) (*ListComplianceRecordsResponse, error)

	// ReportViolation reports a violation
	ReportViolation(ctx context.Context, req *ReportViolationRequest) (*ReportViolationResponse, error)

	// RequestExemption requests an exemption
	RequestExemption(ctx context.Context, req *RequestExemptionRequest) (*RequestExemptionResponse, error)

	// GetComplianceMetrics gets compliance metrics
	GetComplianceMetrics(ctx context.Context, req *GetComplianceMetricsRequest) (*GetComplianceMetricsResponse, error)

	// GenerateComplianceReport generates a compliance report
	GenerateComplianceReport(ctx context.Context, req *GenerateComplianceReportRequest) (*GenerateComplianceReportResponse, error)
}

// GRPCReflectionService defines the interface for gRPC reflection.
type GRPCReflectionService interface {
	// ListServices lists all services
	ListServices(ctx context.Context) ([]*ServiceInfo, error)

	// ListMethods lists all methods of a service
	ListMethods(ctx context.Context, serviceName string) ([]*MethodInfo, error)

	// ListMethodParams lists parameters of a method
	ListMethodParams(ctx context.Context, serviceName, methodName string) ([]*ParameterInfo, error)

	// GetMethodDocumentation gets documentation for a method
	GetMethodDocumentation(ctx context.Context, serviceName, methodName string) (string, error)
}

// StreamHandler defines the interface for handling gRPC streaming.
type StreamHandler interface {
	// HandleServerStreaming handles server-side streaming
	HandleServerStreaming(srv interface{}, ss grpc.ServerStreamingServer[interface{}]) error

	// HandleClientStreaming handles client-side streaming
	HandleClientStreaming(srv interface{}, ss grpc.ClientStreamingServer[interface{}, interface{}]) error

	// HandleBidirectionalStreaming handles bidirectional streaming
	HandleBidirectionalStreaming(srv interface{}, ss grpc.BidiStreamingServer[interface{}, interface{}]) error
}

// GRPCServerOption defines the interface for gRPC server options.
type GRPCServerOption interface {
	// Apply applies the option to the server config
	Apply(config *grpc.ServerConfig)

	// String returns the option name
	String() string
}

// GRPCClientOption defines the interface for gRPC client options.
type GRPCClientOption interface {
	// Apply applies the option to the client config
	Apply(config *grpc.ClientConfig)

	// String returns the option name
	String() string
}

// GRPCLoadBalancer defines the interface for load balancing operations.
type GRPCLoadBalancer interface {
	// GetTarget returns the target address
	GetTarget(ctx context.Context) string

	// Resolve resolves the target address
	Resolve(ctx context.Context) ([]*grpc.Address, error)

	// Notify notifies of address updates
	Notify() <-chan []*grpc.Address

	// Close closes the load balancer
	Close()
}

// GRPCResolver defines the interface for name resolution.
type GRPCResolver interface {
	// Build builds a resolver
	Build(target string, cc grpc.ClientConnInterface, opts grpc.BuildOptions) (grpc.Resolver, error)

	// Scheme returns the resolver scheme
	Scheme() string
}
