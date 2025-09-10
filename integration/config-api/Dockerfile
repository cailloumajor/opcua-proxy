FROM denoland/deno:distroless-2.5.0

WORKDIR /app

COPY main.ts check.ts ./

EXPOSE 3000
CMD ["run", "--allow-net", "main.ts"]
