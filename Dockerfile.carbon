FROM node:20.15.1-alpine3.19 AS builder

WORKDIR /opt/carbon

RUN apk add --update git \
    && rm -rf /var/cache/apk/* \
    && git clone -b 4.9.10 --single-branch https://github.com/carbon-app/carbon.git . \
    && echo "legacy-peer-deps=true" >> .npmrc \
    && yarn install \
    && yarn build \
    && npm prune --production


FROM node:20.15.1-alpine3.19

WORKDIR /opt/carbon

COPY --from=builder /opt/carbon /opt/carbon

CMD ["yarn", "start", "-p", "3000"]
