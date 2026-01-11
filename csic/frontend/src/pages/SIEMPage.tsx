import React from 'react';
import { Box, Typography, Card, CardContent } from '@mui/material';

export const SIEMPage: React.FC = () => {
  return (
    <Box>
      <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
        SIEM Events
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
        Security Information and Event Management monitoring
      </Typography>
      
      <Card>
        <CardContent sx={{ textAlign: 'center', py: 8 }}>
          <Typography variant="h6" sx={{ mb: 2 }}>
            SIEM Dashboard
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Security events and alerts will be displayed here
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
};

export default SIEMPage;
