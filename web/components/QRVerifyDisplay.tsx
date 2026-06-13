"use client";

interface QRVerifyDisplayProps {
  verifyUrl: string;
  token: string;
  onClose: () => void;
}

export default function QRVerifyDisplay({
  verifyUrl,
  token,
  onClose,
}: QRVerifyDisplayProps) {
  return (
    <div
      onClick={onClose}
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 1000,
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: "#fff",
          padding: "2rem",
          borderRadius: "12px",
          textAlign: "center",
          maxWidth: "400px",
          width: "100%",
          boxSizing: "border-box",
        }}
      >
        <h3 style={{ marginTop: 0, marginBottom: "0.5rem" }}>
          Verification QR Code
        </h3>
        <p style={{ fontSize: "0.85rem", color: "#666", marginBottom: "1.5rem" }}>
          Scan this code to verify the credential. Valid for 30 minutes.
        </p>

        {/* QR Code Display Area */}
        <div
          style={{
            width: "256px",
            height: "256px",
            background: "#f8f9fa",
            margin: "0 auto 1.5rem",
            borderRadius: "8px",
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            color: "#999",
            fontSize: "0.85rem",
            border: "2px dashed #ddd",
          }}
        >
          <span style={{ fontSize: "1.5rem", marginBottom: "0.5rem" }}>
            QR
          </span>
          <span>Server-generated QR code</span>
          <span style={{ fontSize: "0.75rem" }}>appears here</span>
        </div>

        {/* Token Display */}
        <div
          style={{
            background: "#f8f9fa",
            padding: "0.5rem",
            borderRadius: "6px",
            marginBottom: "1rem",
            fontSize: "0.75rem",
            color: "#666",
            wordBreak: "break-all",
          }}
        >
          <strong>Token:</strong> {token.substring(0, 16)}...
        </div>

        {/* Verification URL */}
        <p style={{ fontSize: "0.8rem", color: "#999", marginBottom: "1rem" }}>
          Verification URL:{" "}
          <a
            href={verifyUrl}
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: "#3498db" }}
          >
            {verifyUrl.length > 50
              ? verifyUrl.substring(0, 50) + "..."
              : verifyUrl}
          </a>
        </p>

        <button
          onClick={onClose}
          style={{
            padding: "0.5rem 1.5rem",
            background: "#1a1a2e",
            color: "#fff",
            border: "none",
            borderRadius: "6px",
            cursor: "pointer",
            fontWeight: 600,
          }}
        >
          Close
        </button>
      </div>
    </div>
  );
}