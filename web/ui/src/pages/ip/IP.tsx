import { useParams } from "react-router-dom";
import { IPDetailView } from "@/components/ip-detail-view";

const IPPage = () => {
  const { ip } = useParams<{ ip: string }>();
  
  if (!ip) {
    throw new Error("IP address parameter is required");
  }

  return (
    <div className="space-y-6">
      <IPDetailView 
        ipAddress={decodeURIComponent(ip)} 
        onBack={() => window.history.back()} 
      />
    </div>
  );
};

export default IPPage;
