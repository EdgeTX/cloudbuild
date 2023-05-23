import { ConfigProvider, Layout, Result, theme } from "antd";
import { Navigate, Route, Routes } from "react-router-dom";
import { Content } from "antd/es/layout/layout";
import { useColorScheme } from "@hooks/useColorscheme";
import { AuthContext, useAuthenticated } from "@hooks/useAuthenticated";
import Navbar from "@comps/Navbar";
import Footer from "@comps/Footer";
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
        <Layout style={{ minHeight: "100%" }}>
          <Navbar />
          <Content
            className="site-layout"
            style={{ padding: "1rem", paddingBottom: 0 }}
          >
            {isAuth === false && <Result status="error" title={authError} />}
            {isAuth && (
              <Routes>
                <Route path="home" element={<Home />} />
                <Route path="jobs" element={<Jobs />} />
                <Route path="create" element={<JobCreation />} />
                <Route path="workers" element={<Workers />} />
                <Route path="*" element={<Navigate to="home" replace />} />
              </Routes>
            )}
          </Content>
          <Footer />
        </Layout>
      </AuthContext.Provider>
    </ConfigProvider>
  );
}

export default App;
