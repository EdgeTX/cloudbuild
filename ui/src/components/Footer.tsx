import { Footer as AFooter } from "antd/es/layout/layout";

function Footer() {
  return (
    <AFooter
      style={{
        display: "flex",
        height: "1rem",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      Built with ❤️ by the EdgeTX contributors -&nbsp;
      <a href="https://github.com/EdgeTX/cloudbuild">
        source
      </a>
    </AFooter>
  );
}

export default Footer;
