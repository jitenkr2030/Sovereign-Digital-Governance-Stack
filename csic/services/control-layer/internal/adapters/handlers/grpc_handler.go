package handlers

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"csic-platform/control-layer/internal/core/domain"
	"csic-platform/control-layer/internal/core/ports"
	"csic-platform/control-layer/internal/core/services"
	"csic-platform/control-layer/pkg/metrics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// GRPCHandler handles gRPC requests
type GRPCHandler struct {
	policyEngine        services.PolicyEngine
	enforcementHandler  services.EnforcementHandler
	stateRegistry       services.StateRegistry
	interventionService services.InterventionService
	metricsCollector    *metrics.MetricsCollector
	logger              *zap.Logger
	server              *grpc.Server
	healthServer        *health.Server
	mu                  sync.Mutex
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(
	policyEngine services.PolicyEngine,
	enforcementHandler services.EnforcementHandler,
	stateRegistry services.StateRegistry,
	interventionService services.InterventionService,
	metricsCollector *metrics.MetricsCollector,
	logger *zap.Logger,
) *GRPCHandler {
	return &GRPCHandler{
		policyEngine:        policyEngine,
		enforcementHandler:  enforcementHandler,
		stateRegistry:       stateRegistry,
		interventionService: interventionService,
		metricsCollector:    metricsCollector,
		logger:              logger,
		healthServer:        health.NewServer(),
	}
}

// Start starts the gRPC server
func (h *GRPCHandler) Start(port int, logger *zap.Logger) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	addr := ":" + strconv.Itoa(port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server with optional TLS
	var server *grpc.Server

	// Register health server
	grpc_health_v1.RegisterHealthServer(server, h.healthServer)
	reflection.Register(server)

	h.server = server
	h.healthServer.SetServingStatus("control-layer", grpc_health_v1.HealthCheckResponse_SERVING)

	logger.Info("Starting gRPC server", zap.String("addr", addr))

	return server.Serve(lis)
}

// Shutdown gracefully shuts down the gRPC server
func (h *GRPCHandler) Shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.server != nil {
		h.healthServer.SetServingStatus("control-layer", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		h.server.GracefulStop()
	}
}

// PolicyService gRPC service implementation
type PolicyServiceServer struct {
	policyEngine services.PolicyEngine
	metrics      *metrics.MetricsCollector
}

// EnforcementService gRPC service implementation
type EnforcementServiceServer struct {
	enforcementHandler services.EnforcementHandler
	metrics           *metrics.MetricsCollector
}

// StateService gRPC service implementation
type StateServiceServer struct {
	stateRegistry services.StateRegistry
	metrics       *metrics.MetricsCollector
}

// InterventionService gRPC service implementation
type InterventionServiceServer struct {
	interventionService services.InterventionService
	metrics             *metrics.MetricsCollector
}

// ListPolicies lists all policies
func (s *PolicyServiceServer) ListPolicies(ctx context.Context, req *ListPoliciesRequest) (*ListPoliciesResponse, error) {
	start := time.Now()

	policies, err := s.policyEngine.ListPolicies(ctx)
	if err != nil {
		log.Printf("Failed to list policies: %v", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("ListPolicies", "success", duration)

	protoPolicies := make([]*Policy, len(policies))
	for i, p := range policies {
		protoPolicies[i] = domainPolicyToProto(p)
	}

	return &ListPoliciesResponse{
		Policies: protoPolicies,
		Count:    int32(len(policies)),
	}, nil
}

// GetPolicy gets a policy by ID
func (s *PolicyServiceServer) GetPolicy(ctx context.Context, req *GetPolicyRequest) (*Policy, error) {
	start := time.Now()

	policy, err := s.policyEngine.GetPolicy(ctx, req.Id)
	if err != nil {
		s.metrics.RecordGRPCCall("GetPolicy", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("GetPolicy", "success", duration)

	if policy == nil {
		return nil, status.Error(codes.NotFound, "policy not found")
	}

	return domainPolicyToProto(policy), nil
}

// CreatePolicy creates a new policy
func (s *PolicyServiceServer) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*Policy, error) {
	start := time.Now()

	createReq := &domain.CreatePolicyRequest{
		Name:        req.Name,
		Description: req.Description,
		Rule:        protoRuleToDomain(req.Rule),
		Priority:    int(req.Priority),
		IsActive:    req.IsActive,
	}

	policy, err := s.policyEngine.CreatePolicy(ctx, createReq)
	if err != nil {
		s.metrics.RecordGRPCCall("CreatePolicy", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("CreatePolicy", "success", duration)

	return domainPolicyToProto(policy), nil
}

// EvaluatePolicy evaluates a policy
func (s *PolicyServiceServer) EvaluatePolicy(ctx context.Context, req *EvaluatePolicyRequest) (*EvaluatePolicyResponse, error) {
	start := time.Now()

	data := make(map[string]interface{})
	for k, v := range req.Data {
		data[k] = v
	}

	result, err := s.policyEngine.EvaluatePolicy(ctx, req.PolicyId, data)
	if err != nil {
		s.metrics.RecordGRPCCall("EvaluatePolicy", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("EvaluatePolicy", "success", duration)

	return &EvaluatePolicyResponse{
		PolicyId: req.Policy_id,
		Result:   result.Compliant,
		Details:  result.Details,
	}, nil
}

// ListEnforcements lists all enforcements
func (s *EnforcementServiceServer) ListEnforcements(ctx context.Context, req *ListEnforcementsRequest) (*ListEnforcementsResponse, error) {
	start := time.Now()

	enforcements, err := s.enforcementHandler.ListEnforcements(ctx)
	if err != nil {
		s.metrics.RecordGRPCCall("ListEnforcements", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("ListEnforcements", "success", duration)

	protoEnforcements := make([]*Enforcement, len(enforcements))
	for i, e := range enforcements {
		protoEnforcements[i] = domainEnforcementToProto(e)
	}

	return &ListEnforcementsResponse{
		Enforcements: protoEnforcements,
		Count:        int32(len(enforcements)),
	}, nil
}

// GetEnforcement gets an enforcement by ID
func (s *EnforcementServiceServer) GetEnforcement(ctx context.Context, req *GetEnforcementRequest) (*Enforcement, error) {
	start := time.Now()

	enforcement, err := s.enforcementHandler.GetEnforcement(ctx, req.Id)
	if err != nil {
		s.metrics.RecordGRPCCall("GetEnforcement", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("GetEnforcement", "success", duration)

	if enforcement == nil {
		return nil, status.Error(codes.NotFound, "enforcement not found")
	}

	return domainEnforcementToProto(enforcement), nil
}

// CreateEnforcement creates a new enforcement
func (s *EnforcementServiceServer) CreateEnforcement(ctx context.Context, req *CreateEnforcementRequest) (*Enforcement, error) {
	start := time.Now()

	createReq := &domain.CreateEnforcementRequest{
		TargetService: req.TargetService,
		ActionType:    req.ActionType,
		Severity:      req.Severity,
		Message:       req.Message,
		PolicyID:      req.PolicyId,
	}

	enforcement, err := s.enforcementHandler.CreateEnforcement(ctx, createReq)
	if err != nil {
		s.metrics.RecordGRPCCall("CreateEnforcement", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("CreateEnforcement", "success", duration)

	return domainEnforcementToProto(enforcement), nil
}

// UpdateEnforcementStatus updates the status of an enforcement
func (s *EnforcementServiceServer) UpdateEnforcementStatus(ctx context.Context, req *UpdateEnforcementStatusRequest) (*UpdateStatusResponse, error) {
	start := time.Now()

	status, err := domainEnforcementStatusFromString(req.Status)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.enforcementHandler.UpdateStatus(ctx, req.Id, status); err != nil {
		s.metrics.RecordGRPCCall("UpdateEnforcementStatus", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("UpdateEnforcementStatus", "success", duration)

	return &UpdateStatusResponse{Success: true}, nil
}

// GetState gets a state by key
func (s *StateServiceServer) GetState(ctx context.Context, req *GetStateRequest) (*State, error) {
	start := time.Now()

	state, err := s.stateRegistry.GetState(ctx, req.Key)
	if err != nil {
		s.metrics.RecordGRPCCall("GetState", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("GetState", "success", duration)

	return &State{
		Key:   req.Key,
		Value: state.Value,
	}, nil
}

// UpdateState updates a state
func (s *StateServiceServer) UpdateState(ctx context.Context, req *UpdateStateRequest) (*State, error) {
	start := time.Now()

	state, err := s.stateRegistry.UpdateState(ctx, req.Key, req.Value, time.Duration(req.Ttl))
	if err != nil {
		s.metrics.RecordGRPCCall("UpdateState", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("UpdateState", "success", duration)

	return &State{
		Key:   req.Key,
		Value: state.Value,
	}, nil
}

// ListStates lists all states
func (s *StateServiceServer) ListStates(ctx context.Context, req *ListStatesRequest) (*ListStatesResponse, error) {
	start := time.Now()

	states, err := s.stateRegistry.ListStates(ctx)
	if err != nil {
		s.metrics.RecordGRPCCall("ListStates", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("ListStates", "success", duration)

	protoStates := make([]*State, len(states))
	for i, s := range states {
		protoStates[i] = &State{Key: s.Key, Value: s.Value}
	}

	return &ListStatesResponse{States: protoStates}, nil
}

// ListInterventions lists all interventions
func (s *InterventionServiceServer) ListInterventions(ctx context.Context, req *ListInterventionsRequest) (*ListInterventionsResponse, error) {
	start := time.Now()

	interventions, err := s.interventionService.ListInterventions(ctx)
	if err != nil {
		s.metrics.RecordGRPCCall("ListInterventions", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("ListInterventions", "success", duration)

	protoInterventions := make([]*Intervention, len(interventions))
	for i, in := range interventions {
		protoInterventions[i] = domainInterventionToProto(in)
	}

	return &ListInterventionsResponse{
		Interventions: protoInterventions,
		Count:         int32(len(interventions)),
	}, nil
}

// GetIntervention gets an intervention by ID
func (s *InterventionServiceServer) GetIntervention(ctx context.Context, req *GetInterventionRequest) (*Intervention, error) {
	start := time.Now()

	intervention, err := s.interventionService.GetIntervention(ctx, req.Id)
	if err != nil {
		s.metrics.RecordGRPCCall("GetIntervention", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("GetIntervention", "success", duration)

	if intervention == nil {
		return nil, status.Error(codes.NotFound, "intervention not found")
	}

	return domainInterventionToProto(intervention), nil
}

// ResolveIntervention resolves an intervention
func (s *InterventionServiceServer) ResolveIntervention(ctx context.Context, req *ResolveInterventionRequest) (*ResolveInterventionResponse, error) {
	start := time.Now()

	if err := s.interventionService.Resolve(ctx, req.Id, req.Resolution); err != nil {
		s.metrics.RecordGRPCCall("ResolveIntervention", "error", float64(time.Since(start).Milliseconds()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	duration := float64(time.Since(start).Milliseconds())
	s.metrics.RecordGRPCCall("ResolveIntervention", "success", duration)

	return &ResolveInterventionResponse{Success: true}, nil
}

// Helper functions for domain to proto conversion
func domainPolicyToProto(p *domain.Policy) *Policy {
	return &Policy{
		Id:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		Rule:        domainRuleToProto(&p.Rule),
		Priority:    int32(p.Priority),
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt.Unix(),
		UpdatedAt:   p.UpdatedAt.Unix(),
	}
}

func domainRuleToProto(r *domain.PolicyRule) *PolicyRule {
	return &PolicyRule{
		Condition:      r.Condition,
		Target:         r.Target,
		Threshold:      r.Threshold,
		Operator:       r.Operator,
		Action:         r.Action,
		AlertSeverity:  r.AlertSeverity,
	}
}

func protoRuleToDomain(r *PolicyRule) *domain.PolicyRule {
	return &domain.PolicyRule{
		Condition:     r.Condition,
		Target:        r.Target,
		Threshold:     r.Threshold,
		Operator:      r.Operator,
		Action:        r.Action,
		AlertSeverity: r.AlertSeverity,
	}
}

func domainEnforcementToProto(e *domain.Enforcement) *Enforcement {
	return &Enforcement{
		Id:            e.ID.String(),
		PolicyId:      e.PolicyID.String(),
		TargetService: e.TargetService,
		ActionType:    e.ActionType,
		Severity:      e.Severity,
		Status:        string(e.Status),
		Message:       e.Message,
		CreatedAt:     e.CreatedAt.Unix(),
		UpdatedAt:     e.UpdatedAt.Unix(),
	}
}

func domainInterventionToProto(i *domain.Intervention) *Intervention {
	return &Intervention{
		Id:              i.ID.String(),
		PolicyId:        i.PolicyID.String(),
		TargetService:   i.TargetService,
		InterventionType: i.InterventionType,
		Severity:        i.Severity,
		Status:          string(i.Status),
		Reason:          i.Reason,
		StartedAt:       i.StartedAt.Unix(),
		EndedAt:         i.EndedAt.Unix(),
		CreatedAt:       i.CreatedAt.Unix(),
		UpdatedAt:       i.UpdatedAt.Unix(),
	}
}

func domainEnforcementStatusFromString(s string) (domain.EnforcementStatus, error) {
	switch s {
	case "pending":
		return domain.EnforcementStatusPending, nil
	case "in_progress":
		return domain.EnforcementStatusInProgress, nil
	case "completed":
		return domain.EnforcementStatusCompleted, nil
	case "failed":
		return domain.EnforcementStatusFailed, nil
	default:
		return "", fmt.Errorf("unknown enforcement status: %s", s)
	}
}
