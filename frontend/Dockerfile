# Build stage
FROM node:20-alpine AS build
WORKDIR /app

# Accept build arguments for environment variables
ARG VITE_API_BASE_URL
ARG VITE_GOOGLE_CLIENT_ID

# Make them available to the build process
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID

COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Production stage
FROM nginx:stable-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
