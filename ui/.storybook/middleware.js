import { createProxyMiddleware } from "http-proxy-middleware";
import * as dotenv from "dotenv";
dotenv.config();

module.exports = function (app) {
  app.use(
    "/api",
    createProxyMiddleware({
      target: process.env.PROXY,
      changeOrigin: true,
    }),
  );
};
