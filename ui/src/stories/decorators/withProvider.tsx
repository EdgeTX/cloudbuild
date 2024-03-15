import { AuthContext } from "@/hooks/useAuthenticated";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ConfigProvider, MappingAlgorithm, theme } from "antd";
import { BrowserRouter } from "react-router-dom";
import { Decorator } from "@storybook/react";

export const withProvider: Decorator = (Story, options) => {
  const queryClient = new QueryClient();

  const THEMES: Record<string, MappingAlgorithm> = {
    "#F8F8F8": theme.defaultAlgorithm,
    "#333333": theme.darkAlgorithm,
  };

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ConfigProvider
          theme={{
            algorithm: THEMES[options.globals.backgrounds?.value],
          }}
        >
          <AuthContext.Provider
            value={{ token: "hey :)", checkAuth: () => undefined }}
          >
            <Story />
          </AuthContext.Provider>
        </ConfigProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
};
