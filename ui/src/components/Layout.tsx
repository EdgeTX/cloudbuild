import { Layout as AntLayout } from "antd";
import { Content } from "antd/es/layout/layout";
import Navbar from "./Navbar";
import Footer from "./Footer";

interface Props {
  fullHeight?: boolean;
  children: React.ReactNode;
}

function Layout({ fullHeight, children }: Props) {
  return (
    <AntLayout
      style={{
        minHeight: "100%",
        ...(fullHeight ? { maxHeight: "100%", height: "100%" } : {}),
      }}
    >
      <Navbar />
      <Content
        className="site-layout"
        style={{ padding: "1rem", paddingBottom: 0 }}
      >
        {children}
      </Content>
      <Footer />
    </AntLayout>
  );
}

export default Layout;
