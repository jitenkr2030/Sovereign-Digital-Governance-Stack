import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  Box,
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Typography,
  Divider,
  Collapse,
  Avatar,
  Tooltip,
  useTheme,
  useMediaQuery,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Security as SecurityIcon,
  History as HistoryIcon,
  Key as KeyIcon,
  Assessment as AssessmentIcon,
  Settings as SettingsIcon,
  People as PeopleIcon,
  Receipt as ReceiptIcon,
  ExpandLess,
  ExpandMore,
  Shield as ShieldIcon,
  VerifiedUser as VerifiedUserIcon,
} from '@mui/icons-material';
import { useAuth } from '../../context/AuthContext';

const drawerWidth = 280;

interface NavItem {
  label: string;
  path: string;
  icon: React.ReactNode;
  roles?: string[];
  children?: NavItem[];
}

const navItems: NavItem[] = [
  {
    label: 'Dashboard',
    path: '/dashboard',
    icon: <DashboardIcon />,
  },
  {
    label: 'Compliance',
    path: '/compliance',
    icon: <AssessmentIcon />,
  },
  {
    label: 'Audit Logs',
    path: '/audit',
    icon: <HistoryIcon />,
    roles: ['admin', 'regulator', 'auditor'],
  },
  {
    label: 'Chain Verification',
    path: '/audit/verify',
    icon: <VerifiedUserIcon />,
    roles: ['admin', 'regulator', 'auditor'],
  },
  {
    label: 'Key Management',
    path: '/keys',
    icon: <KeyIcon />,
    roles: ['admin', 'auditor'],
  },
  {
    label: 'SIEM Events',
    path: '/siem',
    icon: <SecurityIcon />,
    roles: ['admin', 'regulator'],
  },
  {
    label: 'Licenses',
    path: '/licenses',
    icon: <ReceiptIcon />,
  },
  {
    label: 'Reports',
    path: '/reports',
    icon: <AssessmentIcon />,
    roles: ['admin', 'regulator', 'auditor'],
  },
  {
    label: 'Users',
    path: '/users',
    icon: <PeopleIcon />,
    roles: ['admin'],
  },
  {
    label: 'Settings',
    path: '/settings',
    icon: <SettingsIcon />,
    roles: ['admin'],
  },
];

export const Sidebar: React.FC = () => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [mobileOpen, setMobileOpen] = useState(false);
  const [expandedItems, setExpandedItems] = useState<string[]>([]);
  const { user, hasRole } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleNavClick = (item: NavItem) => {
    if (item.children) {
      setExpandedItems(prev =>
        prev.includes(item.path)
          ? prev.filter(p => p !== item.path)
          : [...prev, item.path]
      );
    } else {
      navigate(item.path);
      if (isMobile) {
        setMobileOpen(false);
      }
    }
  };

  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  const canAccess = (item: NavItem) => {
    if (!item.roles) return true;
    return hasRole(item.roles as any);
  };

  const filteredNavItems = navItems.filter(canAccess);

  const drawer = (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Logo and Brand */}
      <Box
        sx={{
          p: 2.5,
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
          borderBottom: `1px solid ${theme.palette.divider}`,
        }}
      >
        <Box
          sx={{
            width: 40,
            height: 40,
            borderRadius: 2,
            bgcolor: 'primary.main',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <ShieldIcon sx={{ color: 'white', fontSize: 24 }} />
        </Box>
        <Box>
          <Typography variant="h6" sx={{ fontWeight: 700, lineHeight: 1.2 }}>
            CSIC Platform
          </Typography>
          <Typography variant="caption" color="text.secondary">
            Regulatory Dashboard
          </Typography>
        </Box>
      </Box>

      {/* Navigation */}
      <Box sx={{ flex: 1, overflow: 'auto', py: 2 }}>
        <List dense>
          {filteredNavItems.map((item, index) => (
            <React.Fragment key={item.path}>
              {index > 0 && <Divider sx={{ my: 1 }} />}
              <ListItem disablePadding sx={{ px: 1 }}>
                <Tooltip title={isMobile ? item.label : ''} placement="right">
                  <ListItemButton
                    onClick={() => handleNavClick(item)}
                    selected={isActive(item.path)}
                    sx={{
                      borderRadius: 2,
                      mb: 0.5,
                      '&.Mui-selected': {
                        bgcolor: 'primary.main',
                        color: 'white',
                        '&:hover': {
                          bgcolor: 'primary.dark',
                        },
                        '& .MuiListItemIcon-root': {
                          color: 'white',
                        },
                      },
                    }}
                  >
                    <ListItemIcon
                      sx={{
                        minWidth: 40,
                        color: isActive(item.path) ? 'inherit' : 'text.secondary',
                      }}
                    >
                      {item.icon}
                    </ListItemIcon>
                    <ListItemText
                      primary={item.label}
                      primaryTypographyProps={{
                        fontSize: '0.875rem',
                        fontWeight: isActive(item.path) ? 600 : 500,
                      }}
                    />
                    {item.children && (
                      expandedItems.includes(item.path) ? <ExpandLess /> : <ExpandMore />
                    )}
                  </ListItemButton>
                </Tooltip>
              </ListItem>

              {/* Child Items */}
              {item.children && (
                <Collapse in={expandedItems.includes(item.path)} timeout="auto" unmountOnExit>
                  <List dense component="div" disablePadding>
                    {item.children.map((child) => (
                      <ListItem key={child.path} disablePadding sx={{ px: 2 }}>
                        <ListItemButton
                          onClick={() => navigate(child.path)}
                          selected={isActive(child.path)}
                          sx={{
                            borderRadius: 2,
                            pl: 4,
                            '&.Mui-selected': {
                              bgcolor: 'action.selected',
                            },
                          }}
                        >
                          <ListItemIcon sx={{ minWidth: 32 }}>
                            {child.icon}
                          </ListItemIcon>
                          <ListItemText
                            primary={child.label}
                            primaryTypographyProps={{ fontSize: '0.8125rem' }}
                          />
                        </ListItemButton>
                      </ListItem>
                    ))}
                  </List>
                </Collapse>
              )}
            </React.Fragment>
          ))}
        </List>
      </Box>

      {/* User Info */}
      <Box
        sx={{
          p: 2,
          borderTop: `1px solid ${theme.palette.divider}`,
          display: 'flex',
          alignItems: 'center',
          gap: 1.5,
        }}
      >
        <Avatar
          sx={{
            width: 36,
            height: 36,
            bgcolor: 'secondary.main',
            fontSize: '0.875rem',
            fontWeight: 600,
          }}
        >
          {user?.name?.charAt(0) || 'U'}
        </Avatar>
        <Box sx={{ flex: 1, minWidth: 0 }}>
          <Typography
            variant="body2"
            sx={{ fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}
          >
            {user?.name || 'Guest User'}
          </Typography>
          <Typography
            variant="caption"
            color="text.secondary"
            sx={{ textTransform: 'capitalize' }}
          >
            {user?.role || 'viewer'}
          </Typography>
        </Box>
      </Box>
    </Box>
  );

  return (
    <>
      {/* Mobile drawer */}
      <Drawer
        variant="temporary"
        open={mobileOpen}
        onClose={handleDrawerToggle}
        ModalProps={{ keepMounted: true }}
        sx={{
          display: { xs: 'block', md: 'none' },
          '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
        }}
      >
        {drawer}
      </Drawer>

      {/* Desktop drawer */}
      <Drawer
        variant="permanent"
        sx={{
          display: { xs: 'none', md: 'block' },
          '& .MuiDrawer-paper': {
            boxSizing: 'border-box',
            width: drawerWidth,
            borderRight: `1px solid ${theme.palette.divider}`,
          },
        }}
        open
      >
        {drawer}
      </Drawer>
    </>
  );
};

export default Sidebar;
