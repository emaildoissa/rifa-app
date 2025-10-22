# rifa-online-frontend/Dockerfile

# --- Estágio 1: Build ---
# Usamos a imagem oficial do Node.js para construir o React
FROM node:18-alpine AS builder

WORKDIR /app

# Copiamos os package.json e instalamos as dependências
COPY package.json package-lock.json ./
RUN npm install

# Copiamos o resto do código-fonte
COPY . .

# Rodamos o build de produção do Vite
RUN npm run build

# --- Estágio 2: Serve ---
# Usamos a imagem oficial do Nginx (um servidor web leve)
FROM nginx:1.25-alpine

# Copiamos o build estático (a pasta 'dist') do estágio 1 para a pasta web do Nginx
COPY --from=builder /app/dist /usr/share/nginx/html

# Precisamos de uma configuração customizada para o Nginx funcionar com o React Router
# Removemos a config padrão
RUN rm /etc/nginx/conf.d/default.conf

# Copiamos a nossa nova config
COPY nginx.conf /etc/nginx/conf.d/