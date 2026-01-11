// CSIC Platform - Login Page
// Secure authentication page for the regulator dashboard

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login, isLoading } = useAuthStore();
  
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!username || !password) {
      setError('请输入用户名和密码');
      return;
    }

    await login(username, password);
    // Check if login was successful by verifying auth state
    setTimeout(() => {
      if (useAuthStore.getState().isAuthenticated) {
        navigate('/');
      } else {
        setError('用户名或密码错误');
      }
    }, 100);
  };

  const handleDemoLogin = async () => {
    setError(null);
    await login('admin', 'demo');
    // Check if login was successful by verifying auth state
    setTimeout(() => {
      if (useAuthStore.getState().isAuthenticated) {
        navigate('/');
      } else {
        setError('演示登录失败');
      }
    }, 100);
  };

  return (
    <div className="login-page">
      <div className="login-container">
        <div className="login-left">
          <div className="login-branding">
            <div className="csic-logo-large">
              <svg viewBox="0 0 100 100" className="logo-icon-large">
                <circle cx="50" cy="50" r="45" fill="none" stroke="currentColor" strokeWidth="2" />
                <path d="M30 50 L45 65 L70 35" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </div>
            <h1>加密货币国家基础设施承包商</h1>
            <p className="login-subtitle">Crypto State Infrastructure Contractor</p>
            <p className="login-description">
              政府级加密货币监管平台，提供交易所监控、区块链分析、合规管理和风险控制等全面功能。
            </p>
            
            <div className="login-features">
              <div className="feature-item">
                <svg viewBox="0 0 24 24" className="feature-icon">
                  <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
                  <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
                  <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
                </svg>
                <span>实时交易监控</span>
              </div>
              <div className="feature-item">
                <svg viewBox="0 0 24 24" className="feature-icon">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" fill="none" stroke="currentColor" strokeWidth="2" />
                </svg>
                <span>市场操纵检测</span>
              </div>
              <div className="feature-item">
                <svg viewBox="0 0 24 24" className="feature-icon">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
                  <path d="M7 11V7a5 5 0 0 1 10 0v4" fill="none" stroke="currentColor" strokeWidth="2" />
                </svg>
                <span>合规风险管理</span>
              </div>
              <div className="feature-item">
                <svg viewBox="0 0 24 24" className="feature-icon">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
                  <polyline points="14 2 14 8 20 8" fill="none" stroke="currentColor" strokeWidth="2" />
                  <line x1="16" y1="13" x2="8" y2="13" fill="none" stroke="currentColor" strokeWidth="2" />
                  <line x1="16" y1="17" x2="8" y2="17" fill="none" stroke="currentColor" strokeWidth="2" />
                </svg>
                <span>审计追踪报告</span>
              </div>
            </div>
          </div>
        </div>

        <div className="login-right">
          <div className="login-form-container">
            <div className="login-header">
              <h2>欢迎回来</h2>
              <p>请登录您的账户以继续</p>
            </div>

            {error && (
              <div className="login-error">
                <svg viewBox="0 0 24 24" className="error-icon">
                  <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
                  <path d="M12 8v4M12 16h.01" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
                <span>{error}</span>
              </div>
            )}

            <form onSubmit={handleSubmit} className="login-form">
              <div className="form-group">
                <label htmlFor="username" className="form-label">用户名</label>
                <div className="input-wrapper">
                  <svg viewBox="0 0 24 24" className="input-icon">
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" fill="none" stroke="currentColor" strokeWidth="2" />
                    <circle cx="12" cy="7" r="4" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  <input
                    type="text"
                    id="username"
                    className="form-input with-icon"
                    placeholder="请输入用户名"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    autoComplete="username"
                    disabled={isLoading}
                  />
                </div>
              </div>

              <div className="form-group">
                <label htmlFor="password" className="form-label">密码</label>
                <div className="input-wrapper">
                  <svg viewBox="0 0 24 24" className="input-icon">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  <input
                    type={showPassword ? 'text' : 'password'}
                    id="password"
                    className="form-input with-icon"
                    placeholder="请输入密码"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="current-password"
                    disabled={isLoading}
                  />
                  <button
                    type="button"
                    className="password-toggle"
                    onClick={() => setShowPassword(!showPassword)}
                    tabIndex={-1}
                  >
                    {showPassword ? (
                      <svg viewBox="0 0 24 24" className="eye-icon">
                        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" fill="none" stroke="currentColor" strokeWidth="2" />
                        <line x1="1" y1="1" x2="23" y2="23" stroke="currentColor" strokeWidth="2" />
                      </svg>
                    ) : (
                      <svg viewBox="0 0 24 24" className="eye-icon">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                        <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              <div className="form-options">
                <label className="checkbox-wrapper">
                  <input
                    type="checkbox"
                    checked={rememberMe}
                    onChange={(e) => setRememberMe(e.target.checked)}
                  />
                  <span className="checkmark"></span>
                  <span>记住我</span>
                </label>
                <a href="#forgot-password" className="forgot-link">忘记密码？</a>
              </div>

              <button type="submit" className="btn btn-primary btn-lg login-btn" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <span className="loading-spinner-small"></span>
                    <span>登录中...</span>
                  </>
                ) : (
                  '登录'
                )}
              </button>
            </form>

            <div className="login-divider">
              <span>或</span>
            </div>

            <button onClick={handleDemoLogin} className="btn btn-secondary btn-lg demo-btn" disabled={isLoading}>
              使用演示账户登录
            </button>

            <div className="login-footer">
              <p>受保护系统 - 仅限授权人员访问</p>
              <div className="security-badges">
                <span className="security-badge">
                  <svg viewBox="0 0 24 24" className="badge-icon">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  MFA 已启用
                </span>
                <span className="security-badge">
                  <svg viewBox="0 0 24 24" className="badge-icon">
                    <rect x="3" y="11" width="18" height="11" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  256位加密
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
