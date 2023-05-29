import Layout from "@comps/Layout"
import { ConfigProvider, Result, theme } from "antd";
import { Navigate, Route, Routes } from "react-router-dom";
import { useColorScheme } from "@hooks/useColorscheme";
import { AuthContext, useAuthenticated } from "@hooks/useAuthenticated";
import Home from "@pages/Home";
import Jobs from "@pages/Jobs";
import Workers from "@pages/Workers";
import JobCreation from "./pages/JobCreation";

function App() {
  const { isAuth, authError, checkAuth, token } = useAuthenticated();
  const isDarkTheme = useColorScheme();

  return (
    <ConfigProvider
      theme={{
        algorithm: isDarkTheme ? theme.darkAlgorithm : theme.defaultAlgorithm,
      }}
    >
      <AuthContext.Provider value={{ checkAuth, token }}>
        {isAuth === false && (
          <Layout>
            <Result status="error" title={authError} />
          </Layout>
        )}
        {isAuth && (
          <Routes>
            <Route path="home" element={<Home />} />
            <Route path="jobs" element={<Jobs />} />
            <Route path="create" element={<JobCreation />} />
            <Route path="workers" element={<Workers />} />
            <Route path="*" element={<Navigate to="home" replace />} />
          </Routes>
        )}
      </AuthContext.Provider>
    </ConfigProvider>
  );
}

export default App;
