import LegalLayout from '../layout'

export default function PrivacyPolicyPage() {
  return (
    <LegalLayout title="Privacy Policy" effectiveDate="July 25, 2026">
      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>1. Introduction</h2>
      <p className="mb-4 leading-relaxed">
        This Privacy Policy describes how NOANT collects, uses, stores, shares, and protects personal information
        when you use our AI customer support platform ("Service"). We are committed to protecting your privacy and
        handling your data with transparency and care.
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>2. Information We Collect</h2>
      <h3 className="text-lg font-medium mt-6 mb-3" style={{ color: 'var(--text-primary)' }}>2.1 Information You Provide</h3>
      <p className="mb-2 leading-relaxed"><strong>Account Information:</strong> Name, email address, password (hashed), company name, job title, phone number (optional), and billing information.</p>
      <p className="mb-2 leading-relaxed"><strong>Content You Upload:</strong> Knowledge base documents, training materials, conversation templates, brand assets, and custom AI prompts.</p>
      <p className="mb-4 leading-relaxed"><strong>Communications:</strong> Support tickets, feedback, and marketing preferences.</p>

      <h3 className="text-lg font-medium mt-6 mb-3" style={{ color: 'var(--text-primary)' }}>2.2 Information Collected Automatically</h3>
      <p className="mb-2 leading-relaxed"><strong>Usage Data:</strong> Pages visited, features used, actions taken, session duration and frequency.</p>
      <p className="mb-2 leading-relaxed"><strong>Device and Technical Data:</strong> IP address (anonymized after 90 days), browser type, operating system, and referring URL.</p>
      <p className="mb-4 leading-relaxed"><strong>API and Integration Data:</strong> Webhook events, connection status, and message delivery status.</p>

      <h3 className="text-lg font-medium mt-6 mb-3" style={{ color: 'var(--text-primary)' }}>2.3 Information from End Users</h3>
      <p className="mb-4 leading-relaxed">
        When your End Users interact with AI agents deployed through the Service, we may collect conversation data,
        contact information voluntarily provided, and sentiment analysis results. You, as the business operator,
        are the primary Data Controller for End User data.
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>3. How We Use Your Information</h2>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li><strong>Service Provision:</strong> Provide, maintain, and improve the Service; process conversations; generate AI responses; manage Accounts; process payments.</li>
        <li><strong>Communication:</strong> Send Account-related emails; provide support; deliver service updates; send marketing (with consent).</li>
        <li><strong>Improvement:</strong> Analyze usage patterns; develop new features; conduct A/B testing; train AI models (with explicit opt-in consent only).</li>
        <li><strong>Security:</strong> Detect and prevent incidents; enforce Terms; comply with legal obligations; maintain audit logs.</li>
      </ul>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>4. Legal Bases for Processing (GDPR)</h2>
      <div className="overflow-x-auto mb-4">
        <table className="min-w-full text-sm border" style={{ borderColor: 'var(--border-default)' }}>
          <thead>
            <tr style={{ background: 'var(--bg-surface)' }}>
              <th className="text-left p-3 border" style={{ borderColor: 'var(--border-default)', color: 'var(--text-primary)' }}>Legal Basis</th>
              <th className="text-left p-3 border" style={{ borderColor: 'var(--border-default)', color: 'var(--text-primary)' }}>Processing Activity</th>
            </tr>
          </thead>
          <tbody>
            <tr><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Contract Performance</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Providing the Service, managing Accounts, processing payments, delivering support</td></tr>
            <tr style={{ background: 'var(--bg-surface)' }}><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Legitimate Interests</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Improving the Service, ensuring security, preventing fraud, analytics</td></tr>
            <tr><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Consent</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Marketing communications, AI model training, non-essential cookies</td></tr>
            <tr style={{ background: 'var(--bg-surface)' }}><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Legal Obligation</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Responding to lawful requests, maintaining records, tax compliance</td></tr>
          </tbody>
        </table>
      </div>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>5. How We Share Your Information</h2>
      <p className="mb-4 leading-relaxed">
        <strong>We do not sell, rent, or trade your personal information to third parties for their marketing purposes.</strong>
      </p>
      <p className="mb-2 leading-relaxed">We share data with trusted service providers who assist in operating the Service:</p>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li><strong>Hosting provider:</strong> Infrastructure and data storage</li>
        <li><strong>Payment processor:</strong> Billing and payment processing</li>
        <li><strong>AI/LLM providers:</strong> AI response generation (prompt-only, no PII in training)</li>
        <li><strong>Email service:</strong> Transactional emails</li>
        <li><strong>Analytics:</strong> Anonymized usage data</li>
        <li><strong>Error tracking:</strong> Bug monitoring (no PII)</li>
        <li><strong>CDN provider:</strong> Content delivery</li>
      </ul>
      <p className="mb-4 leading-relaxed">
        All service providers are bound by Data Processing Agreements. We may also disclose information if required
        by law, to protect our rights, or in connection with a business transfer.
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>6. Data Retention</h2>
      <div className="overflow-x-auto mb-4">
        <table className="min-w-full text-sm border" style={{ borderColor: 'var(--border-default)' }}>
          <thead>
            <tr style={{ background: 'var(--bg-surface)' }}>
              <th className="text-left p-3 border" style={{ borderColor: 'var(--border-default)', color: 'var(--text-primary)' }}>Data Type</th>
              <th className="text-left p-3 border" style={{ borderColor: 'var(--border-default)', color: 'var(--text-primary)' }}>Retention Period</th>
            </tr>
          </thead>
          <tbody>
            <tr><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Account information</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Duration of Account + 30 days</td></tr>
            <tr style={{ background: 'var(--bg-surface)' }}><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Conversation data</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Duration of Account; deleted within 30 days of closure</td></tr>
            <tr><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Training data</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Duration of Account; deleted within 30 days of closure</td></tr>
            <tr style={{ background: 'var(--bg-surface)' }}><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Payment records</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>7 years (tax law requirement)</td></tr>
            <tr><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Server logs</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>90 days</td></tr>
            <tr style={{ background: 'var(--bg-surface)' }}><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Support tickets</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>2 years after resolution</td></tr>
            <tr><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Backups</td><td className="p-3 border" style={{ borderColor: 'var(--border-default)' }}>Up to 90 days (not individually retrievable)</td></tr>
          </tbody>
        </table>
      </div>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>7. Data Security</h2>
      <h3 className="text-lg font-medium mt-6 mb-3" style={{ color: 'var(--text-primary)' }}>7.1 Technical Measures</h3>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li>TLS 1.2+ encryption for all data in transit</li>
        <li>AES-256 encryption for data at rest</li>
        <li>bcrypt password hashing with salt</li>
        <li>Role-based access control (RBAC)</li>
        <li>Web application firewall and DDoS protection</li>
        <li>Encrypted database connections and parameterized queries</li>
      </ul>
      <h3 className="text-lg font-medium mt-6 mb-3" style={{ color: 'var(--text-primary)' }}>7.2 Organizational Measures</h3>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li>Need-to-know access policies for employees</li>
        <li>Security training and confidentiality agreements</li>
        <li>Regular security audits and penetration testing</li>
        <li>72-hour breach notification procedures</li>
      </ul>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>8. Your Rights</h2>
      <h3 className="text-lg font-medium mt-6 mb-3" style={{ color: 'var(--text-primary)' }}>8.1 Rights Under GDPR (EEA, UK, Switzerland)</h3>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li><strong>Right of Access:</strong> Request a copy of your personal data</li>
        <li><strong>Right to Rectification:</strong> Request correction of inaccurate data</li>
        <li><strong>Right to Erasure:</strong> Request deletion ("Right to be Forgotten")</li>
        <li><strong>Right to Restrict Processing:</strong> Request limitation of processing</li>
        <li><strong>Right to Data Portability:</strong> Receive your data in a machine-readable format</li>
        <li><strong>Right to Object:</strong> Object to processing based on legitimate interests</li>
        <li><strong>Right to Withdraw Consent:</strong> Withdraw consent at any time</li>
        <li><strong>Right to Lodge a Complaint:</strong> File a complaint with your local DPA</li>
      </ul>
      <h3 className="text-lg font-medium mt-6 mb-3" style={{ color: 'var(--text-primary)' }}>8.2 Rights Under CCPA (California)</h3>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li><strong>Right to Know:</strong> Know what personal information is collected and how it's used</li>
        <li><strong>Right to Delete:</strong> Request deletion of personal information</li>
        <li><strong>Right to Opt-Out:</strong> Opt out of sale of personal information (we do not sell)</li>
        <li><strong>Right to Non-Discrimination:</strong> Not be discriminated against for exercising rights</li>
        <li><strong>Right to Correct:</strong> Request correction of inaccurate information</li>
      </ul>
      <p className="mb-4 leading-relaxed">
        To exercise these rights, contact privacy@noant.com or use Settings &gt; Privacy &gt; Data Requests.
        We will respond within 30 days (GDPR) or 45 days (CCPA).
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>9. International Data Transfers</h2>
      <p className="mb-4 leading-relaxed">
        Your data may be processed in countries outside your country of residence. We ensure appropriate safeguards
        through Standard Contractual Clauses (SCCs), adequacy decisions, and Data Processing Agreements with all
        sub-processors. You may request specific data location information by contacting privacy@noant.com.
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>10. Cookies and Tracking</h2>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li><strong>Essential Cookies:</strong> Session management, security, and load balancing — required for the Service</li>
        <li><strong>Analytics Cookies:</strong> Page views, feature usage, and performance (with consent)</li>
      </ul>
      <p className="mb-4 leading-relaxed">
        You can manage cookies through your browser settings and our cookie consent banner. Disabling essential
        cookies may impair Service functionality.
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>11. Children's Privacy</h2>
      <p className="mb-4 leading-relaxed">
        The Service is not directed to individuals under 18. We do not knowingly collect personal information from
        children. If you deploy AI agents that may interact with minors, you are responsible for ensuring compliance
        with COPPA and equivalent laws.
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>12. AI-Specific Privacy Considerations</h2>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li><strong>AI Training:</strong> Your Content trains AI responses exclusively for your Account. We do not use your data for general AI training without explicit opt-in consent.</li>
        <li><strong>AI Limitations:</strong> AI agents may generate inaccurate responses or inadvertently reveal information. We recommend testing before production deployment.</li>
        <li><strong>AI Transparency:</strong> End Users should be informed they are communicating with AI, in compliance with applicable laws (including the EU AI Act where applicable).</li>
      </ul>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>13. Changes to This Policy</h2>
      <p className="mb-4 leading-relaxed">
        We may update this Privacy Policy from time to time. We will notify you via email at least 30 days before
        material changes take effect. Your continued use of the Service after the effective date constitutes
        acceptance of the updated policy.
      </p>

      <h2 className="text-xl font-semibold mt-8 mb-4" style={{ color: 'var(--text-primary)' }}>14. Contact Us</h2>
      <p className="mb-4 leading-relaxed">
        For questions about this Privacy Policy, contact us at:
      </p>
      <ul className="list-disc pl-6 space-y-2 mb-4">
        <li><strong>Email:</strong> privacy@noant.com</li>
        <li><strong>Data Protection Officer:</strong> dpo@noant.com</li>
        <li><strong>Website:</strong> https://noant.com/privacy</li>
      </ul>
    </LegalLayout>
  )
}
