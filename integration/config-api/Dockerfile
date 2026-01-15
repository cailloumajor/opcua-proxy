FROM denoland/deno:distroless-2.6.5

WORKDIR /app

COPY main.ts check.ts ./

EXPOSE 3000
CMD ["run", "--allow-net", "main.ts"]
