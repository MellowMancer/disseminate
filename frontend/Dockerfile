FROM node:lts-alpine

WORKDIR /app
RUN corepack enable

COPY package.json yarn.lock

RUN yarn install --immutable
COPY . .

EXPOSE 5173

CMD ["yarn", "dev"]



