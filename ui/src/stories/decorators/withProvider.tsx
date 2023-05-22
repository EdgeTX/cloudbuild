import React from "react";
import { AuthContext } from "@/hooks/useAuthenticated";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ConfigProvider, theme } from "antd";
import { BrowserRouter } from "react-router-dom";
import { Decorator } from "@storybook/react";

const queryClient = new QueryClient();

const THEMES: Record<string, any> = {
  "#F8F8F8": theme.defaultAlgorithm,
  "#333333": theme.darkAlgorithm,
};

const withProvider: Decorator = (Story, options) => {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ConfigProvider
          theme={{
            algorithm: THEMES[options.globals.backgrounds?.value],
          }}
        >
          <AuthContext.Provider
            value={{ token: "hey :)", checkAuth: () => {} }}
          >
            <Story />
          </AuthContext.Provider>
        </ConfigProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
};

export { withProvider };
