# DKonsole v1.0.7 - Security Hardening Release

## 🔒 Overview

Version 1.0.7 is a **critical security release** that addresses multiple high and critical severity vulnerabilities discovered during a comprehensive security audit. This release implements enterprise-grade security measures while maintaining backward compatibility with existing deployments.

## 🚨 Security Fixes

### Critical Vulnerabilities Addressed

1. **PromQL Injection Prevention** (CRITICAL)
   - **Issue**: User-supplied parameters were directly interpolated into Prometheus queries
   - **Fix**: Implemented strict regex validation (`^[a-zA-Z0-9.-]+$`) for all PromQL parameters
   - **Impact**: Prevents malicious users from executing arbitrary Prometheus queries
   - **Files Modified**: `backend/prometheus.go`, `backend/prometheus_pod.go`

2. **XSS Token Theft via localStorage** (CRITICAL)
   - **Issue**: JWT tokens stored in localStorage were vulnerable to XSS attacks
   - **Fix**: Migrated to HttpOnly cookies with Secure and SameSite=Strict flags
   - **Impact**: Tokens are now inaccessible to JavaScript, preventing XSS-based theft
   - **Files Modified**: `backend/auth.go`, `frontend/src/context/AuthContext.jsx`

3. **Kubernetes Command Injection** (HIGH)
   - **Issue**: Resource names and namespaces were not validated before use in K8s API calls
   - **Fix**: Implemented RFC 1123 DNS subdomain validation for all K8s resource identifiers
   - **Impact**: Prevents injection of malicious commands via crafted resource names
   - **Files Modified**: `backend/handlers.go` (validateK8sName function)

### Security Enhancements

4. **Audit Logging** (NEW)
   - Comprehensive logging of all API requests with user attribution
   - Log format: `[AUDIT] | StatusCode | Duration | User | Method Path`
   - Enables compliance and forensic analysis
   - **Files Added**: `backend/middleware.go`

5. **Rate Limiting** (NEW)
   - Intelligent rate limiting: 300 requests per minute per IP
   - Prevents DoS attacks and brute force attempts
   - Automatically skips WebSocket upgrade requests
   - **Files Added**: `backend/middleware.go`

6. **Content Security Policy Updates**
   - Enhanced CSP to allow Monaco Editor while maintaining security
   - Added `https://cdn.jsdelivr.net` to allowed script/style sources
   - Added `worker-src 'self' blob:` for web workers
   - **Files Modified**: `frontend/nginx.conf`

7. **JWT Secret Enforcement**
   - Strict validation requiring JWT_SECRET environment variable
   - Minimum length requirement: 32 characters
   - Application fails to start if requirements not met
   - **Files Modified**: `backend/auth.go`

## 🐛 Bug Fixes

- **Table Sorting**: Fixed sorting functionality in WorkloadList (was using unsorted array)
- **YAML Editor Loading**: Fixed infinite loading caused by CSP blocking Monaco Editor
- **Secret Display**: Fixed "No data" issue - secrets now properly decode and display
- **Build Error**: Removed duplicate `mux` initialization in main.go

## 🏗️ Architecture Improvements

### Middleware Chain
```
Request → CORS → RateLimit → Audit → Auth → Handler
```

- Unified security policy application across all routes
- Separated public (`/api/login`, `/api/logout`) and secure routes
- Cleaner code with `secure()` and `public()` wrapper functions

### Code Organization
- Created dedicated `backend/middleware.go` for security middleware
- Removed external validation dependency for better portability
- Better separation of concerns (auth, audit, rate limiting)

## 📋 Upgrade Instructions

### From v1.0.6 to v1.0.7

**IMPORTANT**: This release requires environment variable updates.

1. **Update Helm values** (`values.yaml`):
   ```yaml
   # Ensure JWT secret is set and at least 32 characters
   jwtSecret: "your-secret-here-minimum-32-chars"
   
   # Update image tags
   image:
     backend:
       tag: "1.0.7"
     frontend:
       tag: "1.0.7"
   ```

2. **Upgrade the Helm release**:
   ```bash
   helm upgrade dkonsole ./helm/dkonsole -n dkonsole
   ```

3. **Verify deployment**:
   ```bash
   kubectl get pods -n dkonsole
   kubectl logs -n dkonsole deployment/dkonsole-backend
   ```

4. **Clear browser cache** (important for frontend changes):
   - Hard refresh: `Ctrl+Shift+R` (Windows/Linux) or `Cmd+Shift+R` (Mac)
   - Or clear site data in browser DevTools

### Breaking Changes

⚠️ **Authentication Flow**:
- Frontend now uses cookie-based authentication exclusively
- Old localStorage tokens will be ignored
- Users will need to log in again after upgrade

⚠️ **Environment Variables**:
- `JWT_SECRET` is now **required** and must be at least 32 characters
- Application will fail to start if not properly configured

## 🔍 Testing Recommendations

After upgrading, verify:

1. **Authentication**:
   - Log in with admin credentials
   - Verify cookie is set (check DevTools → Application → Cookies)
   - Verify logout clears the cookie

2. **API Functionality**:
   - List resources (Pods, Deployments, etc.)
   - View YAML editor
   - Stream pod logs
   - Execute commands in pods

3. **Security Features**:
   - Check audit logs: `kubectl logs -n dkonsole deployment/dkonsole-backend | grep AUDIT`
   - Verify rate limiting (make >300 requests in 1 minute)
   - Verify secrets display correctly

## 📊 Security Audit Summary

| Severity | Count | Status |
|----------|-------|--------|
| Critical | 2     | ✅ Fixed |
| High     | 1     | ✅ Fixed |
| Medium   | 0     | N/A |
| Low      | 0     | N/A |

## 🔮 Future Roadmap (v1.1.0)

Planned for next major release:
- **Unified Architecture**: Single container serving both frontend and backend
- **Eliminated CORS**: Same-origin policy for enhanced security
- **Simplified Deployment**: One pod, one service, one ingress
- **Better Scalability**: Stateless design for horizontal pod autoscaling

## 📞 Support

If you encounter issues after upgrading:

1. Check the [CHANGELOG.md](./CHANGELOG.md) for detailed changes
2. Review logs: `kubectl logs -n dkonsole deployment/dkonsole-backend`
3. Open an issue on GitHub with:
   - Version information
   - Error logs
   - Steps to reproduce

## 🙏 Acknowledgments

This security release was made possible through:
- Comprehensive security analysis and vulnerability assessment
- Implementation of industry-standard security practices
- Adherence to OWASP guidelines for web application security

---

**Version**: 1.0.7  
**Release Date**: 2025-11-23  
**Docker Images**:
- `dkonsole/dkonsole-backend:1.0.7`
- `dkonsole/dkonsole-frontend:1.0.7`
