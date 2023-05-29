import { useColorScheme } from "@/hooks/useColorscheme";
import { Job } from "@hooks/useJobsData";
import { JsonViewer } from "@textea/json-viewer";

interface Props {
  job: Job;
}

function JobJson({ job }: Props) {
  const isDarkTheme = useColorScheme();
  return (
    <div style={{ maxHeight: "80vh", overflowY: "auto", marginRight: 25 }}>
      <JsonViewer
        value={job}
        theme={isDarkTheme ? "dark" : "light"}
        editable={false}
        displayDataTypes={false}
      />
    </div>
  );
}

export default JobJson;
