import type { Metadata } from "next";
import { Inter } from "next/font/google";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Professional Credentials Wallet",
  description: "Link your RealMe identity to professional licences and verify credentials via QR codes",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={inter.className} style={{ margin: 0, background: "#f5f5f5", minHeight: "100vh" }}>
        <nav
          style={{
            background: "#1a1a2e",
            color: "#fff",
            padding: "1rem 2rem",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <a href="/" style={{ color: "#fff", textDecoration: "none", fontSize: "1.25rem", fontWeight: 700 }}>
            Credentials Wallet
          </a>
          <div style={{ display: "flex", gap: "1rem", alignItems: "center" }}>
            <a href="/dashboard" style={{ color: "#ccc", textDecoration: "none" }}>
              Dashboard
            </a>
          </div>
        </nav>
        <main style={{ maxWidth: "1200px", margin: "0 auto", padding: "2rem" }}>{children}</main>
      </body>
    </html>
  );
}