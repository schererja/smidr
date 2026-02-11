# Multi-stage build for Smidr UI
# Stage 1: Build the React app
FROM node:22-alpine AS build

WORKDIR /app

# Copy package files
COPY package*.json pnpm-lock.yaml ./

# Install dependencies (use pnpm for speed)
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# Copy source files
COPY . .

# Build for production
# Vite will use VITE_API_BASE_URL from environment at build time
RUN pnpm run build

# Stage 2: Serve with nginx
FROM nginx:1.27-alpine

# Copy built assets from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Copy custom nginx config to enable client-side routing and runtime env injection
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
