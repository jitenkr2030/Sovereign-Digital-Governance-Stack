// Approval Inbox Component
// Displays a list of pending approval requests assigned to the user

import React, { useState, useEffect, useMemo } from 'react';
import { 
  Card, 
  CardContent, 
  Typography, 
  Badge, 
  Chip, 
  Button, 
  List, 
  ListItem, 
  ListItemText, 
  ListItemIcon,
  ListItemSecondaryAction,
  IconButton,
  Divider,
  Box,
  CircularProgress,
  Alert,
  Tabs,
  Tab,
  TextField,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
  Tooltip,
} from '@mui/material';
import {
  AccessTime as TimeIcon,
  CheckCircle as ApproveIcon,
  Cancel as RejectIcon,
  Visibility as ViewIcon,
  Notifications as NotifyIcon,
  FilterList as FilterIcon,
  Refresh as RefreshIcon,
  Warning as WarningIcon,
  Person as PersonIcon,
  PriorityHigh as PriorityIcon,
} from '@mui/icons-material';
import { formatDistanceToNow, format } from 'date-fns';
import { useApproval } from '../hooks/useApproval';
import { ApprovalRequest, ApprovalStatus, Priority } from '../types/approval';

interface ApprovalInboxProps {
  userId: string;
  onViewDetails?: (requestId: string) => void;
  onTakeAction?: (requestId: string) => void;
}

const priorityOrder: Record<Priority, number> = {
  critical: 4,
  high: 3,
  normal: 2,
  low: 1,
};

const statusColors: Record<ApprovalStatus, string> = {
  PENDING: '#F59E0B',
  APPROVED: '#10B981',
  REJECTED: '#EF4444',
  EXPIRED: '#6B7280',
  CANCELLED: '#6B7280',
};

export const ApprovalInbox: React.FC<ApprovalInboxProps> = ({
  userId,
  onViewDetails,
  onTakeAction,
}) => {
  const {
    pendingApprovals,
    isLoading,
    error,
    fetchPendingApprovals,
    refresh,
  } = useApproval();

  const [tabValue, setTabValue] = useState(0);
  const [priorityFilter, setPriorityFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [showOverdueOnly, setShowOverdueOnly] = useState(false);

  useEffect(() => {
    if (userId) {
      fetchPendingApprovals(userId);
    }
  }, [userId, fetchPendingApprovals]);

  const filteredRequests = useMemo(() => {
    let filtered = [...pendingApprovals];

    // Filter by tab
    if (tabValue === 0) {
      // Direct assignments
      filtered = filtered.filter(r => !isDelegated(r));
    } else if (tabValue === 1) {
      // Delegated to me
      filtered = filtered.filter(r => isDelegated(r));
    }

    // Filter by priority
    if (priorityFilter !== 'all') {
      filtered = filtered.filter(r => r.priority === priorityFilter);
    }

    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(r => 
        r.resource_id.toLowerCase().includes(query) ||
        r.resource_type.toLowerCase().includes(query) ||
        JSON.stringify(r.context_data).toLowerCase().includes(query)
      );
    }

    // Filter by overdue
    if (showOverdueOnly) {
      filtered = filtered.filter(r => new Date(r.deadline) < new Date());
    }

    // Sort by priority and deadline
    filtered.sort((a, b) => {
      const priorityDiff = priorityOrder[b.priority] - priorityOrder[a.priority];
      if (priorityDiff !== 0) return priorityDiff;
      return new Date(a.deadline).getTime() - new Date(b.deadline).getTime();
    });

    return filtered;
  }, [pendingApprovals, tabValue, priorityFilter, searchQuery, showOverdueOnly]);

  const isDelegated = (request: ApprovalRequest): boolean => {
    // Check if this request was delegated to the current user
    // This would typically come from the backend
    return false;
  };

  const handleRefresh = () => {
    fetchPendingApprovals(userId);
  };

  const getDeadlineStatus = (deadline: string): { status: 'normal' | 'warning' | 'critical'; label: string } => {
    const deadlineDate = new Date(deadline);
    const now = new Date();
    const hoursRemaining = (deadlineDate.getTime() - now.getTime()) / (1000 * 60 * 60);

    if (hoursRemaining < 0) {
      return { status: 'critical', label: 'Overdue' };
    } else if (hoursRemaining < 4) {
      return { status: 'critical', label: `${hoursRemaining.toFixed(1)}h remaining` };
    } else if (hoursRemaining < 24) {
      return { status: 'warning', label: `${hoursRemaining.toFixed(1)}h remaining` };
    }
    return { status: 'normal', label: formatDistanceToNow(deadlineDate, { addSuffix: true }) };
  };

  const getPriorityColor = (priority: Priority): 'error' | 'warning' | 'info' | 'default' => {
    switch (priority) {
      case 'critical': return 'error';
      case 'high': return 'warning';
      case 'normal': return 'info';
      case 'low': return 'default';
    }
  };

  const renderRequestItem = (request: ApprovalRequest) => {
    const deadlineStatus = getDeadlineStatus(request.deadline);
    
    return (
      <React.Fragment key={request.id}>
        <ListItem
          alignItems="flex-start"
          sx={{
            backgroundColor: deadlineStatus.status === 'critical' ? 'rgba(239, 68, 68, 0.05)' : 'transparent',
            borderLeft: `3px solid ${statusColors[request.status]}`,
            '&:hover': {
              backgroundColor: 'rgba(0, 0, 0, 0.02)',
            },
          }}
        >
          <ListItemIcon>
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
              <Chip
                size="small"
                label={request.priority.toUpperCase()}
                color={getPriorityColor(request.priority)}
                icon={request.priority === 'critical' ? <PriorityIcon /> : undefined}
              />
              <TimeIcon
                sx={{
                  mt: 1,
                  color: deadlineStatus.status === 'critical' ? 'error.main' :
                         deadlineStatus.status === 'warning' ? 'warning.main' : 'text.secondary',
                  fontSize: '1.2rem',
                }}
              />
            </Box>
          </ListItemIcon>
          <ListItemText
            primary={
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography variant="subtitle1" component="span">
                  {request.resource_type} Approval
                </Typography>
                <Typography variant="body2" color="text.secondary" component="span">
                  ID: {request.resource_id}
                </Typography>
              </Box>
            }
            secondary={
              <Box component="span">
                <Typography variant="body2" color="text.secondary" component="span" sx={{ display: 'block', mt: 0.5 }}>
                  {deadlineStatus.label}
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
                  <Button
                    size="small"
                    variant="outlined"
                    startIcon={<ViewIcon />}
                    onClick={() => onViewDetails?.(request.id)}
                  >
                    View Details
                  </Button>
                  <Button
                    size="small"
                    variant="contained"
                    color="success"
                    startIcon={<ApproveIcon />}
                    onClick={() => onTakeAction?.(request.id)}
                  >
                    Take Action
                  </Button>
                </Box>
              </Box>
            }
          />
          <ListItemSecondaryAction>
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end' }}>
              <Chip
                size="small"
                label={request.status}
                sx={{ backgroundColor: statusColors[request.status], color: 'white', mb: 1 }}
              />
              <Typography variant="caption" color="text.secondary">
                {format(new Date(request.created_at), 'MMM d, yyyy')}
              </Typography>
            </Box>
          </ListItemSecondaryAction>
        </ListItem>
        <Divider variant="inset" component="li" />
      </React.Fragment>
    );
  };

  if (error) {
    return (
      <Alert severity="error" sx={{ mb: 2 }}>
        {error}
        <Button size="small" onClick={handleRefresh} sx={{ ml: 2 }}>
          Retry
        </Button>
      </Alert>
    );
  }

  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6" component="div">
            Pending Approvals
            <Badge
              badgeContent={filteredRequests.length}
              color="primary"
              sx={{ ml: 2 }}
            />
          </Typography>
          <IconButton onClick={handleRefresh} disabled={isLoading}>
            <RefreshIcon />
          </IconButton>
        </Box>

        <Tabs
          value={tabValue}
          onChange={(_, newValue) => setTabValue(newValue)}
          sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}
        >
          <Tab label="Direct Assignments" />
          <Tab label="Delegated to Me" />
          <Tab label={`All (${pendingApprovals.length})`} />
        </Tabs>

        <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
          <TextField
            size="small"
            placeholder="Search requests..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            sx={{ flex: 1 }}
          />
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>Priority</InputLabel>
            <Select
              value={priorityFilter}
              label="Priority"
              onChange={(e) => setPriorityFilter(e.target.value)}
            >
              <MenuItem value="all">All</MenuItem>
              <MenuItem value="critical">Critical</MenuItem>
              <MenuItem value="high">High</MenuItem>
              <MenuItem value="normal">Normal</MenuItem>
              <MenuItem value="low">Low</MenuItem>
            </Select>
          </FormControl>
          <Button
            variant={showOverdueOnly ? 'contained' : 'outlined'}
            color="warning"
            size="small"
            startIcon={<WarningIcon />}
            onClick={() => setShowOverdueOnly(!showOverdueOnly)}
          >
            Overdue
          </Button>
        </Box>

        {isLoading && !filteredRequests.length ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CircularProgress />
          </Box>
        ) : filteredRequests.length === 0 ? (
          <Box sx={{ textAlign: 'center', p: 4 }}>
            <Typography color="text.secondary">
              No pending approvals matching your criteria
            </Typography>
          </Box>
        ) : (
          <List>
            {filteredRequests.map(renderRequestItem)}
          </List>
        )}
      </CardContent>
    </Card>
  );
};

export default ApprovalInbox;
