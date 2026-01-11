// Approval Detail View Component
// Displays detailed information about an approval request with context and actions

import React, { useState, useEffect } from 'react';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Chip,
  Button,
  Divider,
  Stepper,
  Step,
  StepLabel,
  StepContent,
  Avatar,
  AvatarGroup,
  LinearProgress,
  Alert,
  CircularProgress,
  Paper,
  Grid,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Tabs,
  Tab,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Tooltip,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  CheckCircle as ApproveIcon,
  Cancel as RejectIcon,
  HowToVote as AbstainIcon,
  ExpandMore as ExpandIcon,
  Schedule as TimeIcon,
  Person as PersonIcon,
  Verified as VerifiedIcon,
  Gavel as GavelIcon,
  History as HistoryIcon,
  Description as DocIcon,
  Link as LinkIcon,
  Timer as TimerIcon,
} from '@mui/icons-material';
import { format, formatDistanceToNow } from 'date-fns';
import { useApprovalRequest } from '../hooks/useApproval';
import { ActionType, ApprovalStatus, AuditTrailEntry } from '../types/approval';
import { ConfirmationDialog } from './ConfirmationDialog';

interface ApprovalDetailProps {
  requestId: string;
  currentUserId: string;
  onActionComplete?: (requestId: string, action: ActionType) => void;
}

const statusColors: Record<ApprovalStatus, string> = {
  PENDING: '#F59E0B',
  APPROVED: '#10B981',
  REJECTED: '#EF4444',
  EXPIRED: '#6B7280',
  CANCELLED: '#6B7280',
};

export const ApprovalDetail: React.FC<ApprovalDetailProps> = ({
  requestId,
  currentUserId,
  onActionComplete,
}) => {
  const {
    request,
    approvers,
    summary,
    auditTrail,
    isLoading,
    error,
    refresh,
  } = useApprovalRequest(requestId);

  const [tabValue, setTabValue] = useState(0);
  const [selectedAction, setSelectedAction] = useState<ActionType | null>(null);
  const [comments, setComments] = useState('');
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const canAct = approvers.some(
    a => a.approver_id === currentUserId && !a.has_acted
  );

  const hasActed = approvers.find(
    a => a.approver_id === currentUserId
  )?.has_acted;

  const handleActionClick = (action: ActionType) => {
    setSelectedAction(action);
    setShowConfirmation(true);
  };

  const handleConfirmAction = async () => {
    if (!selectedAction || !currentUserId) return;

    setIsSubmitting(true);
    try {
      // API call would go here
      await refresh();
      onActionComplete?.(requestId, selectedAction);
      setShowConfirmation(false);
      setSelectedAction(null);
      setComments('');
    } catch (err) {
      console.error('Failed to submit action:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancelAction = () => {
    setShowConfirmation(false);
    setSelectedAction(null);
    setComments('');
  };

  const getActionIcon = (action?: ActionType) => {
    switch (action) {
      case 'APPROVE': return <ApproveIcon sx={{ color: 'success.main' }} />;
      case 'REJECT': return <RejectIcon sx={{ color: 'error.main' }} />;
      case 'ABSTAIN': return <AbstainIcon sx={{ color: 'warning.main' }} />;
      default: return <PersonIcon />;
    }
  };

  const getActionColor = (action?: ActionType): 'success' | 'error' | 'warning' | 'default' => {
    switch (action) {
      case 'APPROVE': return 'success';
      case 'REJECT': return 'error';
      case 'ABSTAIN': return 'warning';
      default: return 'default';
    }
  };

  if (isLoading && !request) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{ mb: 2 }}>
        {error}
      </Alert>
    );
  }

  if (!request) {
    return (
      <Alert severity="warning">
        Approval request not found
      </Alert>
    );
  }

  const quorumProgress = summary 
    ? (summary.approvals_count / summary.total_required) * 100 
    : 0;

  return (
    <Box>
      {/* Header */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Box>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 1 }}>
                <Typography variant="h5" component="div">
                  {request.resource_type} Approval Request
                </Typography>
                <Chip
                  label={request.status}
                  sx={{ backgroundColor: statusColors[request.status], color: 'white' }}
                />
              </Box>
              <Typography variant="body2" color="text.secondary">
                Request ID: {request.id}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Resource: {request.resource_type} / {request.resource_id}
              </Typography>
            </Box>
            <Box sx={{ textAlign: 'right' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <TimerIcon color="action" />
                <Typography variant="body2">
                  Deadline: {format(new Date(request.deadline), 'PPP')}
                </Typography>
              </Box>
              <Typography variant="caption" color="text.secondary">
                {formatDistanceToNow(new Date(request.deadline), { addSuffix: true })}
              </Typography>
            </Box>
          </Box>

          {/* Quorum Progress */}
          {summary && (
            <Box sx={{ mt: 3 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">
                  Quorum Progress ({summary.approvals_count}/{summary.total_required})
                </Typography>
                <Typography variant="body2" fontWeight="bold">
                  {summary.quorum_met ? 'Quorum Met!' : 'In Progress'}
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={quorumProgress}
                sx={{
                  height: 10,
                  borderRadius: 5,
                  backgroundColor: 'grey.200',
                  '& .MuiLinearProgress-bar': {
                    backgroundColor: summary.quorum_met ? 'success.main' : 'primary.main',
                  },
                }}
              />
            </Box>
          )}

          {/* Approvers Status */}
          <Box sx={{ mt: 3 }}>
            <Typography variant="subtitle2" gutterBottom>
              Approvers Status
            </Typography>
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
              {approvers.map((approver) => (
                <Tooltip key={approver.id} title={approver.role}>
                  <Chip
                    avatar={
                      <Avatar sx={{ width: 24, height: 24 }}>
                        <PersonIcon />
                      </Avatar>
                    }
                    label={approver.has_acted ? approver.action_type : 'Pending'}
                    color={getActionColor(approver.action_type)}
                    variant={approver.has_acted ? 'filled' : 'outlined'}
                    icon={getActionIcon(approver.action_type)}
                  />
                </Tooltip>
              ))}
            </Box>
          </Box>

          {/* Action Buttons */}
          {canAct && request.status === 'PENDING' && (
            <Box sx={{ mt: 3, display: 'flex', gap: 2 }}>
              <Button
                variant="contained"
                color="success"
                startIcon={<ApproveIcon />}
                onClick={() => handleActionClick('APPROVE')}
              >
                Approve
              </Button>
              <Button
                variant="contained"
                color="error"
                startIcon={<RejectIcon />}
                onClick={() => handleActionClick('REJECT')}
              >
                Reject
              </Button>
              <Button
                variant="outlined"
                color="warning"
                startIcon={<AbstainIcon />}
                onClick={() => handleActionClick('ABSTAIN')}
              >
                Abstain
              </Button>
            </Box>
          )}

          {hasActed && (
            <Alert severity="info" sx={{ mt: 2 }}>
              You have already submitted your decision for this request.
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* Tabs for Details, Context, Audit */}
      <Card>
        <Tabs
          value={tabValue}
          onChange={(_, newValue) => setTabValue(newValue)}
          sx={{ borderBottom: 1, borderColor: 'divider' }}
        >
          <Tab label="Context" icon={<DocIcon />} iconPosition="start" />
          <Tab label="Audit Trail" icon={<HistoryIcon />} iconPosition="start" />
          <Tab label="Details" icon={<GavelIcon />} iconPosition="start" />
        </Tabs>

        <TabPanel value={tabValue} index={0}>
          {/* Context Data */}
          <Box sx={{ p: 2 }}>
            <Typography variant="subtitle2" gutterBottom>
              Request Context
            </Typography>
            <Paper variant="outlined" sx={{ p: 2, backgroundColor: 'grey.50' }}>
              <pre style={{ margin: 0, overflow: 'auto', maxHeight: 400 }}>
                {JSON.stringify(request.context_data, null, 2)}
              </pre>
            </Paper>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          {/* Audit Trail */}
          <Box sx={{ p: 2 }}>
            <Typography variant="subtitle2" gutterBottom>
              Audit Trail
            </Typography>
            <List>
              {auditTrail.map((entry, index) => (
                <AuditTrailItem key={entry.action_id} entry={entry} isFirst={index === 0} />
              ))}
            </List>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          {/* Details */}
          <Box sx={{ p: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" color="text.secondary">
                  Requested By
                </Typography>
                <Typography variant="body1">{request.requester_id}</Typography>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" color="text.secondary">
                  Created At
                </Typography>
                <Typography variant="body1">
                  {format(new Date(request.created_at), 'PPP p')}
                </Typography>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" color="text.secondary">
                  Priority
                </Typography>
                <Chip label={request.priority.toUpperCase()} size="small" />
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="subtitle2" color="text.secondary">
                  Workflow ID
                </Typography>
                <Typography variant="body1">{request.workflow_id}</Typography>
              </Grid>
            </Grid>
          </Box>
        </TabPanel>
      </Card>

      {/* Confirmation Dialog */}
      <ConfirmationDialog
        isOpen={showConfirmation}
        title={`${selectedAction} Confirmation`}
        message={`Are you sure you want to ${selectedAction?.toLowerCase()} this request?`}
        actionType={selectedAction || 'APPROVE'}
        requestInfo={{
          id: request.id,
          resource_type: request.resource_type,
          resource_id: request.resource_id,
        }}
        comments={comments}
        onCommentsChange={setComments}
        onConfirm={handleConfirmAction}
        onCancel={handleCancelAction}
        isLoading={isSubmitting}
      />
    </Box>
  );
};

// Tab Panel Component
const TabPanel: React.FC<{ children: React.ReactNode; value: number; index: number }> = ({
  children,
  value,
  index,
}) => (
  <div role="tabpanel" hidden={value !== index}>
    {value === index && children}
  </div>
);

// Audit Trail Item Component
const AuditTrailItem: React.FC<{ entry: AuditTrailEntry; isFirst: boolean }> = ({
  entry,
  isFirst,
}) => {
  const getActionColor = (action: string): 'success' | 'error' | 'warning' => {
    switch (action) {
      case 'APPROVE': return 'success';
      case 'REJECT': return 'error';
      default: return 'warning';
    }
  };

  return (
    <ListItem alignItems="flex-start">
      <ListItemAvatar>
        <Avatar sx={{ backgroundColor: `${getActionColor(entry.action_type)}.main` }}>
          {getActionIcon(entry.action_type)}
        </Avatar>
      </ListItemAvatar>
      <ListItemText
        primary={
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography variant="subtitle2">
              {entry.actor_name}
            </Typography>
            <Chip
              label={entry.action_type}
              size="small"
              color={getActionColor(entry.action_type)}
            />
            {entry.delegated_from && (
              <Chip
                label="Delegated"
                size="small"
                variant="outlined"
              />
            )}
          </Box>
        }
        secondary={
          <>
            <Typography component="span" variant="body2" color="text.primary">
              {entry.comments || 'No comments'}
            </Typography>
            <br />
            <Typography component="span" variant="caption" color="text.secondary">
              {format(new Date(entry.timestamp), 'PPP p')}
            </Typography>
            {entry.ip_address && (
              <Typography component="span" variant="caption" color="text.secondary">
                {' '}from {entry.ip_address}
              </Typography>
            )}
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 0.5 }}>
              {entry.signature_valid ? (
                <Chip
                  icon={<VerifiedIcon />}
                  label="Cryptographically Verified"
                  size="small"
                  color="success"
                  variant="outlined"
                />
              ) : (
                <Chip
                  label="Signature Invalid"
                  size="small"
                  color="error"
                  variant="outlined"
                />
              )}
            </Box>
          </>
        }
      />
    </ListItem>
  );
};

const getActionIcon = (action: string) => {
  switch (action) {
    case 'APPROVE': return <ApproveIcon sx={{ color: 'success.main' }} />;
    case 'REJECT': return <RejectIcon sx={{ color: 'error.main' }} />;
    case 'ABSTAIN': return <AbstainIcon sx={{ color: 'warning.main' }} />;
    default: return <PersonIcon />;
  }
};

export default ApprovalDetail;
