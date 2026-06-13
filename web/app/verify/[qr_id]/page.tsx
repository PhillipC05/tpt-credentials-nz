"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";

interface VerificationResult {
  valid: boolean;
  full_name?: string;
  professional?: string;
  licence_number?: string;
  status?: string;
  verified_at?: string;
  expires_at?: string;
  error?: string;
}

export default function VerifyPage() {
  const params = useParams();
  const qrId = params.qr_id as string;

  const [loading, setLoading] = useState(true);
  const [result, setResult] = useState<VerificationResult | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!qrId) return;

    const verify = async () => {
      try {
        const res = await fetch(`/api/verify/${qrId}`);
        const data = await res.json();
        setResult(data);
      } catch {
        setError("Failed to verify credential. Please try again.");
      } finally {
        setLoading(false);
      }
    };

    verify();
  }, [qrId]);

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: "4rem 1rem" }}>
        <div
          style={{
            width: "48px",
            height: "48px",
            border: "4px solid #e0e0e0",
            borderTopColor: "#1a1a2e",
            borderRadius: "50%",
            animation: "spin 1s linear infinite",
            margin: "0 auto 1rem",
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
        <p style={{ color: "#666" }}>Verifying credential...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ maxWidth: "500px", margin: "4rem auto", textAlign: "center" }}>
        <div
          style={{
            background: "#fde8e8",
            padding: "2rem",
            borderRadius: "12px",
          }}
        >
          <h2 style={{ color: "#c0392b", marginTop: 0 }}>Verification Error</h2>
          <p style={{ color: "#666" }}>{error}</p>
          <a
            href="/"
            style={{ color: "#3498db", textDecoration: "underline" }}
          >
            Go to Home
          </a>
        </div>
      </div>
    );
  }

  if (!result || !result.valid) {
    return (
      <div style={{ maxWidth: "500px", margin: "4rem auto", textAlign: "center" }}>
        <div
          style={{
            background: "#fff3cd",
            padding: "2rem",
            borderRadius: "12px",
            border: "1px solid #ffc107",
          }}
        >
          <h2 style={{ color: "#856404", marginTop: 0 }}>Credential Not Verified</h2>
          <p style={{ color: "#666" }}>{result?.error || "The credential could not be verified. The QR code may have expired or already been used."}</p>
          <div style={{ marginTop: "1rem", fontSize: "0.85rem", color: "#999" }}>
            <p>Possible reasons:</p>
            <ul style={{ textAlign: "left", maxWidth: "300px", margin: "0 auto" }}>
              <li>QR code has expired (valid for 30 minutes)</li>
              <li>QR code has already been scanned</li>
              <li>Invalid or tampered QR code</li>
            </ul>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: "500px", margin: "4rem auto" }}>
      <div
        style={{
          background: "#e8fde8",
          padding: "0.75rem",
          borderRadius: "8px",
          textAlign: "center",
          marginBottom: "1.5rem",
          color: "#27ae60",
          fontWeight: 600,
        }}
      >
        Credential Verified Successfully
      </div>

      <div
        style={{
          background: "#fff",
          padding: "2rem",
          borderRadius: "12px",
          boxShadow: "0 2px 12px rgba(0,0,0,0.08)",
        }}
      >
        <div style={{ textAlign: "center", marginBottom: "1.5rem" }}>
          <div
            style={{
              width: "64px",
              height: "64px",
              background: "#27ae60",
              borderRadius: "50%",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              margin: "0 auto 1rem",
              fontSize: "2rem",
              color: "#fff",
            }}
          >
            ✓
          </div>
          <h2 style={{ margin: 0, fontSize: "1.25rem" }}>{result.full_name}</h2>
        </div>

        <div style={{ borderTop: "1px solid #eee", paddingTop: "1rem" }}>
          <div style={{ marginBottom: "0.75rem" }}>
            <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Professional Body</span>
            <span style={{ fontWeight: 600 }}>{result.professional}</span>
          </div>
          <div style={{ marginBottom: "0.75rem" }}>
            <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Licence Number</span>
            <span style={{ fontWeight: 600 }}>{result.licence_number}</span>
          </div>
          <div style={{ marginBottom: "0.75rem" }}>
            <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Status</span>
            <span
              style={{
                display: "inline-block",
                padding: "0.15rem 0.5rem",
                borderRadius: "12px",
                fontSize: "0.8rem",
                fontWeight: 600,
                background: "#27ae6020",
                color: "#27ae60",
              }}
            >
              {result.status?.toUpperCase()}
            </span>
          </div>
          {result.verified_at && (
            <div style={{ marginBottom: "0.75rem" }}>
              <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Verified At</span>
              <span>{new Date(result.verified_at).toLocaleString()}</span>
            </div>
          )}
          {result.expires_at && (
            <div>
              <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Expires At</span>
              <span>{new Date(result.expires_at).toLocaleString()}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}