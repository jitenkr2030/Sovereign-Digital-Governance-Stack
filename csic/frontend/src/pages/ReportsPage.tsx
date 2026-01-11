import React from 'react';
import { Box, Typography, Card, CardContent, Button } from '@mui/material';
import { Assessment as AssessmentIcon, Add as AddIcon } from '@mui/icons-material';

export const ReportsPage: React.FC = () => {
  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Box>
          <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
            Reports
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Generate and download compliance reports
          </Typography>
        </Box>
        <Button variant="contained" startIcon={<AddIcon />}>
          Generate Report
        </Button>
      </Box>
      
      <Card>
        <CardContent sx={{ textAlign: 'center', py: 8 }}>
          <AssessmentIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
          <Typography variant="h6" sx={{ mb: 2 }}>
            Compliance Reports
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Generate and export regulatory compliance reports
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
};

export default ReportsPage;
