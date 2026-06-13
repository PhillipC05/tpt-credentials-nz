"use client";

interface Credential {
  id: string;
  user_id: string;
  professional_body_id: string;
  licence_number: string;
  full_name: string;
  status: string;
  verified_at: string | null;
  expires_at: string | null;
  last_checked_at: string | null;
  created_at: string;
  updated_at: string;
}

interface CredentialCardProps {
  credential: Credential;
  onRefresh: (id: string) => void;
  onRevoke: (id: string) => void;
  onGenerateQR: (id: string) => void;
}

const getStatusColor = (status: string): string => {
  switch (status) {
    case "active":
      return "#27ae60";
    case "pending":
      return "#f39c12";
    case "revoked":
      return "#e74c3c";
    case "expired":
      return "#95a5a6";
    default:
      return "#666";
  }
};

export default function CredentialCard({
  credential,
  onRefresh,
  onRevoke,
  onGenerateQR,
}: CredentialCardProps) {
  return (
    <div
      style={{
        background: "#fff",
        padding: "1.25rem",
        borderRadius: "12px",
        boxShadow: "0 2px 8px rgba(0,0,0,0.06)",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <div>
        <div style={{ fontWeight: 600, marginBottom: "0.25rem" }}>
          {credential.full_name}
        </div>
        <div style={{ fontSize: "0.85rem", color: "#666" }}>
          Licence: {credential.licence_number}
        </div>
        <div style={{ marginTop: "0.5rem" }}>
          <span
            style={{
              display: "inline-block",
              padding: "0.15rem 0.5rem",
              borderRadius: "12px",
              fontSize: "0.75rem",
              fontWeight: 600,
              background: getStatusColor(credential.status) + "20",
              color: getStatusColor(credential.status),
            }}
          >
            {credential.status.toUpperCase()}
          </span>
          {credential.last_checked_at && (
            <span style={{ fontSize: "0.7rem", color: "#999", marginLeft: "0.5rem" }}>
              Last checked: {new Date(credential.last_checked_at).toLocaleDateString()}
            </span>
          )}
        </div>
      </div>

      <div style={{ display: "flex", gap: "0.5rem" }}>
        <button
          onClick={() => onRefresh(credential.id)}
          style={{
            padding: "0.35rem 0.75rem",
            background: "#3498db",
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
            fontSize: "0.8rem",
            fontWeight: 500,
          }}
        >
          Refresh
        </button>
        <button
          onClick={() => onGenerateQR(credential.id)}
          style={{
            padding: "0.35rem 0.75rem",
            background: "#2ecc71",
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
            fontSize: "0.8rem",
            fontWeight: 500,
          }}
        >
          QR Code
        </button>
        {credential.status !== "revoked" && (
          <button
            onClick={() => onRevoke(credential.id)}
            style={{
              padding: "0.35rem 0.75rem",
              background: "#e74c3c",
              color: "#fff",
              border: "none",
              borderRadius: "6px",
              cursor: "pointer",
              fontSize: "0.8rem",
              fontWeight: 500,
            }}
          >
            Revoke
          </button>
        )}
      </div>
    </div>
  );
}