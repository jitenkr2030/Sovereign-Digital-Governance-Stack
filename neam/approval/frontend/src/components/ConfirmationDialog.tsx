// Confirmation Dialog Component
// Secure action confirmation dialog for approval decisions

import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Box,
  Typography,
  Alert,
  CircularProgress,
  Divider,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Checkbox,
  FormControlLabel,
} from '@mui/material';
import {
  CheckCircle as ApproveIcon,
  Cancel as RejectIcon,
  HowToVote as AbstainIcon,
  Warning as WarningIcon,
  Info as InfoIcon,
  Security as SecurityIcon,
} from '@mui/icons-material';
import { ActionType } from '../types/approval';

interface ConfirmationDialogProps {
  isOpen: boolean;
  title: string;
  message: string;
  actionType: ActionType;
  requestInfo: {
    id: string;
    resource_type: string;
    resource_id: string;
  };
  comments: string;
  onCommentsChange: (comments: string) => void;
  onConfirm: (comments: string) => void;
  onCancel: () => void;
  isLoading?: boolean;
  requireConfirmation?: boolean;
  showSecurityNotice?: boolean;
}

const actionConfig: Record<ActionType, { color: 'success' | 'error' | 'warning'; icon: React.ReactNode; label: string }> = {
  APPROVE: {
    color: 'success',
    icon: <ApproveIcon />,
    label: 'Approve',
  },
  REJECT: {
    color: 'error',
    icon: <RejectIcon />,
    label: 'Reject',
  },
  ABSTAIN: {
    color: 'warning',
    icon: <AbstainIcon />,
    label: 'Abstain',
  },
};

export const ConfirmationDialog: React.FC<ConfirmationDialogProps> = ({
  isOpen,
  title,
  message,
  actionType,
  requestInfo,
  comments,
  onCommentsChange,
  onConfirm,
  onCancel,
  isLoading = false,
  requireConfirmation = true,
  showSecurityNotice = true,
}) => {
  const [confirmText, setConfirmText] = React.useState('');
  const [acknowledgeRisks, setAcknowledgeRisks] = React.useState(false);
  const config = actionConfig[actionType];

  const canConfirm = () => {
    if (!requireConfirmation) return true;
    if (actionType === 'REJECT' && !acknowledgeRisks) return false;
    return confirmText.toUpperCase() === config.label.toUpperCase();
  };

  const handleConfirm = () => {
    onConfirm(comments);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && canConfirm() && !isLoading) {
      handleConfirm();
    }
  };

  return (
    <Dialog
      open={isOpen}
      onClose={onCancel}
      maxWidth="sm"
      fullWidth
      onKeyPress={handleKeyPress}
    >
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Box
            sx={{
              color: `${config.color}.main`,
              display: 'flex',
            }}
          >
            {config.icon}
          </Box>
          <Typography variant="h6">{title}</Typography>
        </Box>
      </DialogTitle>

      <DialogContent dividers>
        {/* Request Info */}
        <Box
          sx={{
            backgroundColor: 'grey.50',
            p: 2,
            borderRadius: 1,
            mb: 2,
          }}
        >
          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
            Request Details
          </Typography>
          <Typography variant="body1">
            <strong>Type:</strong> {requestInfo.resource_type}
          </Typography>
          <Typography variant="body1">
            <strong>ID:</strong> {requestInfo.resource_id}
          </Typography>
        </Box>

        {/* Warning for Reject */}
        {actionType === 'REJECT' && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            Rejecting this request will prevent it from being approved. This action cannot be easily reversed.
          </Alert>
        )}

        {/* Security Notice */}
        {showSecurityNotice && (
          <Alert
            severity="info"
            icon={<SecurityIcon />}
            sx={{ mb: 2 }}
          >
            Your decision will be cryptographically signed and recorded in the immutable audit ledger.
            This provides non-repudiation for your action.
          </Alert>
        )}

        {/* Confirmation Message */}
        <Typography variant="body1" sx={{ mb: 2 }}>
          {message}
        </Typography>

        {/* Comments */}
        <TextField
          fullWidth
          multiline
          rows={3}
          label="Comments (optional)"
          placeholder="Add any comments or justification for your decision..."
          value={comments}
          onChange={(e) => onCommentsChange(e.target.value)}
          sx={{ mb: 2 }}
        />

        {/* Confirmation for Reject */}
        {actionType === 'REJECT' && requireConfirmation && (
          <Box sx={{ mb: 2 }}>
            <FormControlLabel
              control={
                <Checkbox
                  checked={acknowledgeRisks}
                  onChange={(e) => setAcknowledgeRisks(e.target.checked)}
                  color={config.color}
                />
              }
              label="I understand that rejecting this request cannot be easily reversed and may have significant consequences."
            />
          </Box>
        )}

        {/* Type-to-confirm for sensitive actions */}
        {actionType !== 'ABSTAIN' && requireConfirmation && (
          <Box>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
              Type <strong>{config.label}</strong> to confirm:
            </Typography>
            <TextField
              fullWidth
              size="small"
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              placeholder={`Type "${config.label}" here`}
              error={confirmText !== '' && !canConfirm()}
              helperText={
                confirmText !== '' && !canConfirm()
                  ? `Type exactly "${config.label}" to confirm`
                  : ''
              }
            />
          </Box>
        )}

        {/* Audit Notice */}
        <Box
          sx={{
            mt: 2,
            p: 2,
            backgroundColor: 'info.50',
            borderRadius: 1,
            border: '1px solid',
            borderColor: 'info.light',
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
            <InfoIcon color="info" fontSize="small" />
            <Typography variant="subtitle2">
              This action will be recorded with:
            </Typography>
          </Box>
          <List dense sx={{ py: 0 }}>
            <ListItem sx={{ py: 0.5 }}>
              <ListItemText
                primary="Timestamp"
                secondary={new Date().toISOString()}
                primaryTypographyProps={{ variant: 'body2' }}
                secondaryTypographyProps={{ variant: 'caption' }}
              />
            </ListItem>
            <ListItem sx={{ py: 0.5 }}>
              <ListItemText
                primary="IP Address"
                secondary="Your current IP will be recorded"
                primaryTypographyProps={{ variant: 'body2' }}
                secondaryTypographyProps={{ variant: 'caption' }}
              />
            </ListItem>
            <ListItem sx={{ py: 0.5 }}>
              <ListItemText
                primary="Cryptographic Signature"
                secondary="SHA-256 hash of your decision"
                primaryTypographyProps={{ variant: 'body2' }}
                secondaryTypographyProps={{ variant: 'caption' }}
              />
            </ListItem>
          </List>
        </Box>
      </DialogContent>

      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button
          onClick={onCancel}
          disabled={isLoading}
          color="inherit"
        >
          Cancel
        </Button>
        <Button
          onClick={handleConfirm}
          variant="contained"
          color={config.color}
          disabled={!canConfirm() || isLoading}
          startIcon={isLoading ? <CircularProgress size={20} color={config.color} /> : config.icon}
        >
          {isLoading ? 'Processing...' : `${config.label} Request`}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default ConfirmationDialog;
