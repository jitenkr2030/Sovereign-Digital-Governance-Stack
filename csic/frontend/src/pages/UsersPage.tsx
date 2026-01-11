import React from 'react';
import { Box, Typography, Card, CardContent, Button } from '@mui/material';
import { People as PeopleIcon, Add as AddIcon } from '@mui/icons-material';

export const UsersPage: React.FC = () => {
  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 4 }}>
        <Box>
          <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
            User Management
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Manage platform users and access permissions
          </Typography>
        </Box>
        <Button variant="contained" startIcon={<AddIcon />}>
          Add User
        </Button>
      </Box>
      
      <Card>
        <CardContent sx={{ textAlign: 'center', py: 8 }}>
          <PeopleIcon sx={{ fontSize: 64, color: 'text.secondary', mb: 2 }} />
          <Typography variant="h6" sx={{ mb: 2 }}>
            User Administration
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Manage user accounts, roles, and permissions
          </Typography>
        </CardContent>
      </Card>
    </Box>
  );
};

export default UsersPage;
