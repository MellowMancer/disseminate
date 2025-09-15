FROM node:lts-alpine

WORKDIR /app
RUN corepack enable

COPY package.json yarn.lock ./

RUN yarn install --immutable
COPY . .

COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh

EXPOSE 5173

ENTRYPOINT ["./entrypoint.sh"]



