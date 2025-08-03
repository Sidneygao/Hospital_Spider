# Hospital Spider Deployment Guide

## Render Deployment

### Prerequisites
1. GitHub repository must be public OR Render must have access to your private repository
2. Valid AMAP API key
3. Render account

### Step 1: Repository Setup

#### Option A: Make Repository Public (Recommended)
1. Go to https://github.com/Sidneygao/Hospital_Spider
2. Click "Settings" → "General" → "Danger Zone"
3. Click "Change repository visibility"
4. Select "Make public"
5. Confirm the change

#### Option B: Configure Private Repository Access
1. In Render dashboard, go to "Settings" → "Repository"
2. Connect your GitHub account
3. Ensure Render has access to your repository

### Step 2: Environment Variables Setup

#### Backend Service
In Render dashboard, set these environment variables:

```
AMAP_API_KEY=your_actual_amap_api_key
PORT=8080
GIN_MODE=release
```

#### Frontend Service
In Render dashboard, set these environment variables:

```
REACT_APP_API_URL=https://your-backend-service-name.onrender.com
```

### Step 3: Deploy Services

#### Backend Deployment
1. Create new Web Service in Render
2. Connect to your GitHub repository
3. Configure:
   - **Name**: `hospital-spider-backend`
   - **Environment**: `Go`
   - **Build Command**: `cd backend && go build -o main .`
   - **Start Command**: `cd backend && ./main`
   - **Health Check Path**: `/api/hospitals`

#### Frontend Deployment
1. Create new Static Site in Render
2. Connect to your GitHub repository
3. Configure:
   - **Name**: `hospital-spider-frontend`
   - **Build Command**: `cd frontend && npm install && npm run build`
   - **Publish Directory**: `frontend/build`

### Step 4: Update Frontend API URL

After backend deployment, update the frontend environment variable:
```
REACT_APP_API_URL=https://hospital-spider-backend.onrender.com
```

### Step 5: Verify Deployment

1. **Backend Health Check**: Visit `https://your-backend-url.onrender.com/api/hospitals`
2. **Frontend**: Visit your frontend URL
3. **Test Search**: Try searching for hospitals near a location

## Alternative: Using render.yaml

If you prefer using the `render.yaml` file:

1. Push the `render.yaml` file to your repository
2. In Render dashboard, create a new "Blueprint"
3. Connect to your repository
4. Render will automatically create both services

## Troubleshooting

### Common Issues

1. **Repository Access Error**
   - Make repository public OR
   - Ensure Render has GitHub access

2. **Build Failures**
   - Check Go version compatibility
   - Verify all dependencies are in go.mod
   - Check Node.js version for frontend

3. **Environment Variables**
   - Ensure AMAP_API_KEY is set correctly
   - Verify REACT_APP_API_URL points to correct backend URL

4. **CORS Issues**
   - Backend is configured to allow all origins in production
   - Check if frontend is calling correct backend URL

### Debug Commands

```bash
# Check backend logs
curl https://your-backend-url.onrender.com/api/hospitals

# Check frontend build
cd frontend && npm run build

# Test local development
cd backend && go run main.go
cd frontend && npm start
```

## Local Development

### Backend
```bash
cd backend
go mod tidy
go run main.go
```

### Frontend
```bash
cd frontend
npm install
npm start
```

## Production Considerations

1. **Database**: Consider upgrading from SQLite to PostgreSQL for production
2. **Caching**: Implement Redis for better performance
3. **Monitoring**: Add logging and monitoring
4. **Security**: Implement proper authentication and rate limiting
5. **SSL**: Render provides automatic SSL certificates

## Support

For issues with:
- **Render**: Check Render documentation and support
- **GitHub**: Ensure repository permissions are correct
- **AMAP API**: Verify API key and quotas 