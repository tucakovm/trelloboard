FROM node:18-alpine


WORKDIR /app


COPY package.json package-lock.json ./



RUN npm install && npm install -g @angular/cli


COPY . .


EXPOSE 4200

# Start the Angular development server
CMD ["ng", "serve", "--host","0.0.0.0"]
