FROM node:20.1-alpine

WORKDIR /usr/src

COPY package.json ./

# hadolint ignore=DL3060
RUN yarn install --no-lockfile

EXPOSE 3000
CMD ["yarn", "run", "json-server", "--host", "0.0.0.0", "--read-only", "config.json"]
