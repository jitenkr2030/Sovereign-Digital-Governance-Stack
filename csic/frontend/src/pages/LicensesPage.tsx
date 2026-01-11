import React from 'react';
import { Box, Typography, Card, CardContent } from '@mui/material';
import { Receipt as ReceiptIcon } from '@mui/icons-material';

export const LicensesPage: React.FC = () => {
  return (
    <Box>
      <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
        Licenses
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
        Manage platform licenses and certifications
      </Typography>
      
      <Card>
        <CardContent sx={{ textAlign: 'center', py: 8 }}>
          <ReceiptIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
          <Typography variant="h6" sx={{ mb: 2 }}>
            License Management
          </Typography>
          <Typography variant="body2" color="text.secondary">
            View and manage your platform licenses
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
};

export default LicensesPage;
