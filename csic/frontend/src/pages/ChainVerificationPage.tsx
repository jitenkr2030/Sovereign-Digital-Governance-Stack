import React, { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Grid,
  Alert,
  CircularProgress,
  Chip,
  Paper,
  Stepper,
  Step,
  StepLabel,
  Divider,
} from '@mui/material';
import {
  VerifiedUser as VerifiedUserIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Security as SecurityIcon,
} from '@mui/icons-material';
import { api } from '../services/api';

interface VerificationResult {
  isValid: boolean;
  startId: string;
  endId: string;
  entriesVerified: number;
  verifiedAt: string;
  errors?: string[];
}

export const ChainVerificationPage: React.FC = () => {
  const [startId, setStartId] = useState('');
  const [endId, setEndId] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<VerificationResult | null>(null);
  const [error, setError] = useState('');

  const handleVerify = async () => {
    if (!startId || !endId) {
      setError('Please enter both start and end entry IDs');
      return;
    }

    setLoading(true);
    setError('');
    setResult(null);

    try {
      const response = await api.security.verifyChain(startId, endId);
      setResult(response.data);
    } catch (err: any) {
      setError(err.message || 'Verification failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box>
      {/* Page Header */}
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
          Chain Verification
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Verify the integrity of the tamper-proof hash chain for audit logs
        </Typography>
      </Box>

      <Grid container spacing={4}>
        {/* Verification Form */}
        <Grid item xs={12} lg={5}>
          <Card>
            <CardContent sx={{ p: 4 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
                <Box
                  sx={{
                    width: 48,
                    height: 48,
                    borderRadius: 2,
                    bgcolor: 'primary.main',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <VerifiedUserIcon sx={{ color: 'white' }} />
                </Box>
                <Box>
                  <Typography variant="h6" sx={{ fontWeight: 600 }}>
                    Verify Hash Chain
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Validate the cryptographic chain between two audit entries
                  </Typography>
                </Box>
              </Box>

              {error && (
                <Alert severity="error" sx={{ mb: 3 }}>
                  {error}
                </Alert>
              )}

              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
                <TextField
                  fullWidth
                  label="Start Entry ID"
                  placeholder="e.g., log-1735209600000-0"
                  value={startId}
                  onChange={(e) => setStartId(e.target.value)}
                  helperText="The beginning of the chain segment to verify"
                />

                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Divider sx={{ width: '100%' }}>
                    <Chip size="small" label="TO" />
                  </Divider>
                </Box>

                <TextField
                  fullWidth
                  label="End Entry ID"
                  placeholder="e.g., log-1735296000000-0"
                  value={endId}
                  onChange={(e) => setEndId(e.target.value)}
                  helperText="The end of the chain segment to verify"
                />

                <Button
                  variant="contained"
                  size="large"
                  onClick={handleVerify}
                  disabled={loading || !startId || !endId}
                  startIcon={loading ? <CircularProgress size={20} color="inherit" /> : <SecurityIcon />}
                  sx={{ mt: 2 }}
                >
                  {loading ? 'Verifying...' : 'Verify Chain Integrity'}
                </Button>
              </Box>

              {/* Sample Data */}
              <Box sx={{ mt: 4, p: 2, bgcolor: 'background.default', borderRadius: 2 }}>
                <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                  Sample Entry IDs (from Audit Logs):
                </Typography>
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                  {['log-1735209600000-0', 'log-1735224000000-1', 'log-1735238400000-2'].map((id) => (
                    <Button
                      key={id}
                      size="small"
                      variant="text"
                      onClick={() => setStartId(id)}
                      sx={{ justifyContent: 'flex-start', textTransform: 'none' }}
                    >
                      {id}
                    </Button>
                  ))}
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Verification Result */}
        <Grid item xs={12} lg={7}>
          {result ? (
            <Card>
              <CardContent sx={{ p: 4 }}>
                <Box sx={{ textAlign: 'center', mb: 4 }}>
                  <Box
                    sx={{
                      width: 80,
                      height: 80,
                      borderRadius: '50%',
                      bgcolor: result.isValid ? 'success.main' : 'error.main',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mx: 'auto',
                      mb: 2,
                    }}
                  >
                    {result.isValid ? (
                      <CheckCircleIcon sx={{ color: 'white', fontSize: 48 }} />
                    ) : (
                      <ErrorIcon sx={{ color: 'white', fontSize: 48 }} />
                    )}
                  </Box>
                  <Typography variant="h5" sx={{ fontWeight: 700 }}>
                    {result.isValid ? 'Chain Verified' : 'Verification Failed'}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    {result.isValid
                      ? 'The hash chain integrity has been confirmed'
                      : 'Tampering detected in the chain'}
                  </Typography>
                </Box>

                {/* Chain Visualization */}
                <Paper variant="outlined" sx={{ p: 3, mb: 3 }}>
                  <Typography variant="subtitle2" color="text.secondary" sx={{ mb: 2 }}>
                    Chain Segment
                  </Typography>
                  <Stepper orientation="vertical" activeStep={0}>
                    <Step completed>
                      <StepLabel
                        optional={
                          <Typography variant="caption" color="text.secondary">
                            ID: {result.startId}
                          </Typography>
                        }
                      >
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Typography variant="body2" sx={{ fontWeight: 600 }}>
                            Start Entry
                          </Typography>
                          <Chip size="small" label="Verified" color="success" />
                        </Box>
                      </StepLabel>
                    </Step>
                    <Step>
                      <StepLabel
                        optional={
                          <Typography variant="caption" color="text.secondary">
                            {result.entriesVerified} entries verified
                          </Typography>
                        }
                      >
                        <Typography variant="body2">
                          Intermediate Entries
                        </Typography>
                      </StepLabel>
                    </Step>
                    <Step completed>
                      <StepLabel
                        optional={
                          <Typography variant="caption" color="text.secondary">
                            ID: {result.endId}
                          </Typography>
                        }
                      >
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Typography variant="body2" sx={{ fontWeight: 600 }}>
                            End Entry
                          </Typography>
                          <Chip size="small" label="Verified" color="success" />
                        </Box>
                      </StepLabel>
                    </Step>
                  </Stepper>
                </Paper>

                {/* Verification Details */}
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Paper variant="outlined" sx={{ p: 2, textAlign: 'center' }}>
                      <Typography variant="h4" sx={{ fontWeight: 700, color: 'success.main' }}>
                        {result.entriesVerified}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Entries Verified
                      </Typography>
                    </Paper>
                  </Grid>
                  <Grid item xs={6}>
                    <Paper variant="outlined" sx={{ p: 2, textAlign: 'center' }}>
                      <Typography variant="h4" sx={{ fontWeight: 700 }}>
                        SHA-256
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Algorithm Used
                      </Typography>
                    </Paper>
                  </Grid>
                </Grid>

                {/* Verification Timestamp */}
                <Box sx={{ mt: 3, p: 2, bgcolor: 'background.default', borderRadius: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    Verification Timestamp
                  </Typography>
                  <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                    {new Date(result.verifiedAt).toLocaleString()}
                  </Typography>
                </Box>

                {result.errors && result.errors.length > 0 && (
                  <Alert severity="error" sx={{ mt: 3 }}>
                    <Typography variant="subtitle2">Errors Detected:</Typography>
                    <ul style={{ margin: '8px 0', paddingLeft: 20 }}>
                      {result.errors.map((err, i) => (
                        <li key={i}>{err}</li>
                      ))}
                    </ul>
                  </Alert>
                )}
              </CardContent>
            </Card>
          ) : (
            <Card sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <CardContent sx={{ textAlign: 'center', py: 8 }}>
                <VerifiedUserIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
                <Typography variant="h6" sx={{ fontWeight: 600, mb: 1 }}>
                  Ready to Verify
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Enter the start and end entry IDs to verify the integrity of the audit log chain.
                </Typography>
              </CardContent>
            </Card>
          )}
        </Grid>
      </Grid>

      {/* How It Works Section */}
      <Card sx={{ mt: 4 }}>
        <CardContent sx={{ p: 4 }}>
          <Typography variant="h6" sx={{ fontWeight: 600, mb: 3 }}>
            How Chain Verification Works
          </Typography>
          <Grid container spacing={4}>
            <Grid item xs={12} md={4}>
              <Box sx={{ textAlign: 'center' }}>
                <Box
                  sx={{
                    width: 48,
                    height: 48,
                    borderRadius: '50%',
                    bgcolor: 'primary.main',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    mx: 'auto',
                    mb: 2,
                  }}
                >
                  <Typography variant="h6" sx={{ color: 'white', fontWeight: 700 }}>
                    1
                  </Typography>
                </Box>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                  Hash Chain Structure
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Each audit log entry contains a cryptographic hash that links to the previous entry, creating an immutable chain.
                </Typography>
              </Box>
            </Grid>
            <Grid item xs={12} md={4}>
              <Box sx={{ textAlign: 'center' }}>
                <Box
                  sx={{
                    width: 48,
                    height: 48,
                    borderRadius: '50%',
                    bgcolor: 'primary.main',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    mx: 'auto',
                    mb: 2,
                  }}
                >
                  <Typography variant="h6" sx={{ color: 'white', fontWeight: 700 }}>
                    2
                  </Typography>
                </Box>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                  SHA-256 Hashing
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Every entry is hashed using SHA-256, ensuring that any modification to historical data would break the chain.
                </Typography>
              </Box>
            </Grid>
            <Grid item xs={12} md={4}>
              <Box sx={{ textAlign: 'center' }}>
                <Box
                  sx={{
                    width: 48,
                    height: 48,
                    borderRadius: '50%',
                    bgcolor: 'primary.main',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    mx: 'auto',
                    mb: 2,
                  }}
                >
                  <Typography variant="h6" sx={{ color: 'white', fontWeight: 700 }}>
                    3
                  </Typography>
                </Box>
                <Typography variant="subtitle1" sx={{ fontWeight: 600, mb: 1 }}>
                  Verification Process
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  By verifying hashes from start to end, we can confirm that no tampering has occurred within the specified range.
                </Typography>
              </Box>
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    </Box>
  );
};

export default ChainVerificationPage;
