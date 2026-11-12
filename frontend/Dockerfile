FROM node:24-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm install

# Default "/api" -> dilayani nginx (lihat nginx.conf), jadi tidak bergantung pada IP.
ARG VITE_API_URL=/api
ENV VITE_API_URL=$VITE_API_URL

COPY . .
RUN npm run build

FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose port 80
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
