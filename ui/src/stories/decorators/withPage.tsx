import Footer from "@/components/Footer";
import Navbar from "@/components/Navbar";
import { Decorator } from "@storybook/react";
import { Layout } from "antd";
import { Content } from "antd/es/layout/layout";

const withPage: Decorator = (Story) => (
  <Layout style={{ minHeight: "100%" }}>
    <Navbar />
    <Content
      className="site-layout"
      style={{ padding: "1rem", paddingBottom: 0 }}
    >
      <Story />
    </Content>
    <Footer />
  </Layout>
)

export { withPage };
