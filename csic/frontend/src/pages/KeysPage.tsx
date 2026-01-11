import React from 'react';
import { Box, Typography, Card, CardContent, Button } from '@mui/material';
import { Key as KeyIcon, Add as AddIcon } from '@mui/icons-material';

export const KeysPage: React.FC = () => {
  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Box>
          <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
            Key Management
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage cryptographic keys and certificates
          </Typography>
        </Box>
        <Button variant="contained" startIcon={<AddIcon />}>
          Generate New Key
        </Button>
      </Box>
      
      <Card>
        <CardContent sx={{ textAlign: 'center', py: 8 }}>
          <KeyIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
          <Typography variant="h6" sx={{ mb: 2 }}>
            Cryptographic Keys
          </Typography>
          <Typography variant="body2" color="text.secondary">
            View and manage your encryption and signing keys
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
};

export default KeysPage;
