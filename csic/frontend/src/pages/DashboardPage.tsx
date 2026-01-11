import React, { useEffect, useState } from 'react';
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  Chip,
  Avatar,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  IconButton,
  Skeleton,
  useTheme,
  Button,
} from '@mui/material';
import {
  Security as SecurityIcon,
  Warning as WarningIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Key as KeyIcon,
  Receipt as ReceiptIcon,
  Refresh as RefreshIcon,
  ArrowForward as ArrowForwardIcon,
} from '@mui/icons-material';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  color: string;
  loading?: boolean;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  subtitle,
  icon,
  color,
  loading,
}) => {

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent sx={{ p: 3 }}>
        {loading ? (
          <>
            <Skeleton variant="text" width="60%" height={24} />
            <Skeleton variant="text" width="40%" height={48} />
            <Skeleton variant="text" width="80%" height={20} />
          </>
        ) : (
          <>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
              <Box
                sx={{
                  width: 44,
                  height: 44,
                  borderRadius: 2,
                  bgcolor: `${color}15`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: color,
                }}
              >
                {icon}
              </Box>
            </Box>
            <Typography variant="h4" sx={{ fontWeight: 700, mb: 0.5 }}>
              {value}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {title}
            </Typography>
            {subtitle && (
              <Typography variant="caption" color="text.secondary">
                {subtitle}
              </Typography>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
};

interface ComplianceGaugeProps {
  score: number;
  loading?: boolean;
}

const ComplianceGauge: React.FC<ComplianceGaugeProps> = ({ score, loading }) => {
  const theme = useTheme();
  
  const getColor = (value: number) => {
    if (value >= 90) return theme.palette.success.main;
    if (value >= 70) return theme.palette.warning.main;
    return theme.palette.error.main;
  };

  const data = [
    { name: 'Score', value: score },
    { name: 'Remaining', value: 100 - score },
  ];

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent sx={{ p: 3, textAlign: 'center' }}>
        <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>
          Compliance Score
        </Typography>
        {loading ? (
          <Skeleton variant="circular" width={160} height={160} sx={{ mx: 'auto' }} />
        ) : (
          <Box sx={{ position: 'relative', display: 'inline-flex' }}>
            <ResponsiveContainer width={160} height={160}>
              <PieChart>
                <Pie
                  data={data}
                  cx="50%"
                  cy="50%"
                  innerRadius={50}
                  outerRadius={70}
                  startAngle={180}
                  endAngle={0}
                  dataKey="value"
                  stroke="none"
                >
                  <Cell fill={getColor(score)} />
                  <Cell fill={theme.palette.grey[200]} />
                </Pie>
              </PieChart>
            </ResponsiveContainer>
            <Box
              sx={{
                position: 'absolute',
                inset: 0,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexDirection: 'column',
              }}
            >
              <Typography variant="h3" sx={{ fontWeight: 700, color: getColor(score) }}>
                {score}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                out of 100
              </Typography>
            </Box>
          </Box>
        )}
        <Box sx={{ mt: 2 }}>
          <Chip
            size="small"
            icon={<CheckCircleIcon />}
            label="Good Standing"
            color="success"
          />
        </Box>
      </CardContent>
    </Card>
  );
};

interface AlertItem {
  id: string;
  type: 'warning' | 'error' | 'info';
  title: string;
  message: string;
  time: string;
}

const RecentAlerts: React.FC = () => {
  const theme = useTheme();
  const [loading, setLoading] = useState(true);
  const [alerts, setAlerts] = useState<AlertItem[]>([]);

  useEffect(() => {
    const fetchAlerts = async () => {
      try {
        const response = await api.security.getSIEMEvents({ pageSize: 5 });
        const events = response.items || [];
        setAlerts(
          events.map((e: any) => ({
            id: e.id,
            type: e.severity === 'critical' ? 'error' : e.severity === 'high' ? 'warning' : 'info',
            title: e.eventType,
            message: e.action,
            time: new Date(e.timestamp).toLocaleTimeString(),
          }))
        );
      } finally {
        setLoading(false);
      }
    };

    fetchAlerts();
  }, []);

  const getIcon = (type: AlertItem['type']) => {
    switch (type) {
      case 'error':
        return <ErrorIcon sx={{ color: theme.palette.error.main }} />;
      case 'warning':
        return <WarningIcon sx={{ color: theme.palette.warning.main }} />;
      default:
        return <SecurityIcon sx={{ color: theme.palette.info.main }} />;
    }
  };

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            Recent Alerts
          </Typography>
          <IconButton size="small">
            <RefreshIcon />
          </IconButton>
        </Box>
        {loading ? (
          [...Array(5)].map((_, i) => (
            <Box key={i} sx={{ display: 'flex', gap: 2, mb: 2 }}>
              <Skeleton variant="circular" width={32} height={32} />
              <Box sx={{ flex: 1 }}>
                <Skeleton variant="text" width="70%" />
                <Skeleton variant="text" width="50%" />
              </Box>
            </Box>
          ))
        ) : (
          <List dense disablePadding>
            {alerts.map((alert) => (
              <ListItem
                key={alert.id}
                disablePadding
                sx={{ mb: 1.5, bgcolor: 'background.default', borderRadius: 2, p: 1.5 }}
                secondaryAction={
                  <IconButton edge="end" size="small">
                    <ArrowForwardIcon fontSize="small" />
                  </IconButton>
                }
              >
                <ListItemAvatar>
                  <Avatar sx={{ bgcolor: 'background.default', width: 36, height: 36 }}>
                    {getIcon(alert.type)}
                  </Avatar>
                </ListItemAvatar>
                <ListItemText
                  primary={alert.title}
                  secondary={alert.message}
                  primaryTypographyProps={{ fontWeight: 600, fontSize: '0.875rem' }}
                  secondaryTypographyProps={{ fontSize: '0.75rem' }}
                />
              </ListItem>
            ))}
          </List>
        )}
      </CardContent>
    </Card>
  );
};

interface ChartDataPoint {
  time: string;
  value: number;
}

const ComplianceTrendChart: React.FC = () => {
  const theme = useTheme();
  const [data, setData] = useState<ChartDataPoint[]>([]);

  useEffect(() => {
    const now = Date.now();
    const points: ChartDataPoint[] = [];
    for (let i = 30; i >= 0; i--) {
      const date = new Date(now - i * 24 * 60 * 60 * 1000);
      points.push({
        time: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        value: 85 + Math.random() * 10,
      });
    }
    setData(points);
  }, []);

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent sx={{ p: 3 }}>
        <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>
          Compliance Trend
        </Typography>
        <ResponsiveContainer width="100%" height={200}>
          <AreaChart data={data}>
            <defs>
              <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor={theme.palette.primary.main} stopOpacity={0.3} />
                <stop offset="95%" stopColor={theme.palette.primary.main} stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke={theme.palette.divider} />
            <XAxis
              dataKey="time"
              tick={{ fontSize: 11 }}
              tickLine={false}
              axisLine={false}
              interval={7}
            />
            <YAxis
              domain={[70, 100]}
              tick={{ fontSize: 11 }}
              tickLine={false}
              axisLine={false}
            />
            <RechartsTooltip
              contentStyle={{
                backgroundColor: theme.palette.background.paper,
                border: `1px solid ${theme.palette.divider}`,
                borderRadius: 8,
              }}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke={theme.palette.primary.main}
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#colorValue)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
};

export const DashboardPage: React.FC = () => {
  const theme = useTheme();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<any>(null);
  const [complianceScore, setComplianceScore] = useState(0);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const statsRes = await api.security.getDashboardStats();
        const complianceRes = await api.security.getComplianceScore();
        setStats(statsRes);
        setComplianceScore(complianceRes.overall || 0);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  return (
    <Box>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ fontWeight: 700, mb: 1 }}>
          Dashboard
        </Typography>
        <Typography variant="body1" color="text.secondary">
          Overview of your regulatory compliance status and security metrics
        </Typography>
      </Box>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} lg={3}>
          <StatCard
            title="Total Users"
            value={stats?.totalUsers || 0}
            subtitle="Active platform users"
            icon={<SecurityIcon />}
            color={theme.palette.primary.main}
            loading={loading}
          />
        </Grid>
        <Grid item xs={12} sm={6} lg={3}>
          <StatCard
            title="Active Licenses"
            value={stats?.activeLicenses || 0}
            subtitle="Valid licenses"
            icon={<ReceiptIcon />}
            color={theme.palette.success.main}
            loading={loading}
          />
        </Grid>
        <Grid item xs={12} sm={6} lg={3}>
          <StatCard
            title="Pending Alerts"
            value={stats?.pendingAlerts || 0}
            subtitle="Requires attention"
            icon={<WarningIcon />}
            color={theme.palette.warning.main}
            loading={loading}
          />
        </Grid>
        <Grid item xs={12} sm={6} lg={3}>
          <StatCard
            title="Keys Expiring"
            value={stats?.keysExpiringSoon || 0}
            subtitle="Within 30 days"
            icon={<KeyIcon />}
            color={theme.palette.error.main}
            loading={loading}
          />
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={4}>
          <ComplianceGauge score={complianceScore} loading={loading} />
        </Grid>
        <Grid item xs={12} md={8}>
          <ComplianceTrendChart />
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} lg={6}>
          <RecentAlerts />
        </Grid>
        <Grid item xs={12} lg={6}>
          <Card sx={{ height: '100%' }}>
            <CardContent sx={{ p: 3 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                <Typography variant="h6" sx={{ fontWeight: 600 }}>
                  Quick Actions
                </Typography>
              </Box>
              <Grid container spacing={2}>
                {[
                  { label: 'View Audit Logs', path: '/audit', color: theme.palette.primary.main },
                  { label: 'Verify Chain', path: '/audit/verify', color: theme.palette.secondary.main },
                  { label: 'Manage Keys', path: '/keys', color: theme.palette.warning.main },
                  { label: 'Generate Report', path: '/reports', color: theme.palette.success.main },
                ].map((action) => (
                  <Grid item xs={6} key={action.path}>
                    <Button
                      fullWidth
                      variant="outlined"
                      onClick={() => navigate(action.path)}
                      sx={{
                        py: 2,
                        justifyContent: 'flex-start',
                        borderColor: action.color,
                        color: action.color,
                        '&:hover': {
                          borderColor: action.color,
                          bgcolor: `${action.color}10`,
                        },
                      }}
                    >
                      {action.label}
                    </Button>
                  </Grid>
                ))}
              </Grid>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default DashboardPage;
