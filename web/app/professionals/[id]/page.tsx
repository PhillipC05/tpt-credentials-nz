"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";

interface PublicProfile {
  id: string;
  professional: string;
  full_name: string;
  licence_number: string;
  status: string;
  verified_at: string | null;
  expires_at: string | null;
}

export default function ProfessionalPage() {
  const params = useParams();
  const id = params.id as string;

  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!id) return;

    const fetchProfile = async () => {
      try {
        const res = await fetch(`/api/public/professionals/${id}`);
        if (!res.ok) {
          setError("Professional not found");
          return;
        }
        const data = await res.json();
        setProfile(data);
      } catch {
        setError("Failed to load profile");
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, [id]);

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
        <p style={{ color: "#666" }}>Loading professional profile...</p>
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div style={{ maxWidth: "500px", margin: "4rem auto", textAlign: "center" }}>
        <div
          style={{
            background: "#fde8e8",
            padding: "2rem",
            borderRadius: "12px",
          }}
        >
          <h2 style={{ color: "#c0392b", marginTop: 0 }}>Not Found</h2>
          <p style={{ color: "#666" }}>{error || "This professional profile could not be found."}</p>
        </div>
      </div>
    );
  }

  const getStatusColor = (status: string) => {
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

  return (
    <div style={{ maxWidth: "600px", margin: "2rem auto" }}>
      <div
        style={{
          background: "#fff",
          padding: "2rem",
          borderRadius: "12px",
          boxShadow: "0 2px 12px rgba(0,0,0,0.08)",
        }}
      >
        <div style={{ textAlign: "center", marginBottom: "2rem" }}>
          <div
            style={{
              width: "80px",
              height: "80px",
              background: "#1a1a2e",
              borderRadius: "50%",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              margin: "0 auto 1rem",
              fontSize: "2rem",
              color: "#fff",
              fontWeight: 700,
            }}
          >
            {profile.full_name.charAt(0)}
          </div>
          <h1 style={{ margin: 0, fontSize: "1.5rem" }}>{profile.full_name}</h1>
          <p style={{ color: "#666", margin: "0.25rem 0 0" }}>{profile.professional}</p>
        </div>

        <div style={{ borderTop: "1px solid #eee", paddingTop: "1.5rem" }}>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "1rem",
            }}
          >
            <div>
              <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Licence Number</span>
              <span style={{ fontWeight: 600 }}>{profile.licence_number}</span>
            </div>
            <div>
              <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Status</span>
              <span
                style={{
                  display: "inline-block",
                  padding: "0.15rem 0.5rem",
                  borderRadius: "12px",
                  fontSize: "0.8rem",
                  fontWeight: 600,
                  background: getStatusColor(profile.status) + "20",
                  color: getStatusColor(profile.status),
                }}
              >
                {profile.status.toUpperCase()}
              </span>
            </div>
            {profile.verified_at && (
              <div>
                <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Verified At</span>
                <span>{new Date(profile.verified_at).toLocaleDateString()}</span>
              </div>
            )}
            {profile.expires_at && (
              <div>
                <span style={{ fontSize: "0.8rem", color: "#999", display: "block" }}>Expires At</span>
                <span>{new Date(profile.expires_at).toLocaleDateString()}</span>
              </div>
            )}
          </div>
        </div>

        <div
          style={{
            marginTop: "2rem",
            padding: "1rem",
            background: "#f8f9fa",
            borderRadius: "8px",
            textAlign: "center",
            fontSize: "0.85rem",
            color: "#999",
          }}
        >
          This professional's credentials are verified through the TPT Credentials Wallet.
          Information is provided directly from the register of{" "}
          <strong>{profile.professional}</strong>.
        </div>
      </div>
    </div>
  );
}