# ---- Build Stage ----
FROM node:25-alpine AS builder

WORKDIR /app

# Install dependencies
# RUN npm install @vitejs/plugin-react --save-dev
COPY package.json ./
RUN npm install

# Copy source
COPY . .

# Build the app
RUN npm run build


# ---- Production Stage ----
FROM nginx:1.29-alpine

# Remove default nginx config
RUN rm -rf /usr/share/nginx/html/*

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
