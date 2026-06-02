import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import {
  ArrowRight, Sparkles, MessageSquare, GraduationCap, BarChart3,
  Link2, Bot, Zap, Check, Globe, Shield, LayoutGrid,
  Menu, X, Code2
} from 'lucide-react'

// --- Types for mock playground chat ---
interface ChatMessage {
  id: string
  sender: 'ai' | 'user'
  text: string
  timestamp: string
}

export default function LandingPage() {
  const { user } = useAuth()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [isAnnual, setIsAnnual] = useState(false)

  // AI Playground states
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: '1',
      sender: 'ai',
      text: 'Hi! I am the Noant AI assistant. Click one of the questions below or type yours to see how I handle customer queries using custom trained Q&A.',
      timestamp: 'Just now'
    }
  ])
  const [inputValue, setInputValue] = useState('')
  const [isTyping, setIsTyping] = useState(false)

  const presetQuestions = [
    { label: 'How does human takeover work?', answer: 'If Noant AI encounters a question with low confidence (below 0.65 similarity), it flags the chat as "Escalated". Support agents are instantly notified via WebSockets, and they can take over the conversation immediately from the dashboard.' },
    { label: 'Can I connect my WhatsApp number?', answer: 'Yes! Noant supports a self-hosted OpenWA WhatsApp container. You scan a QR code from the dashboard, and your WhatsApp conversations are instantly routed through Noant in real time.' },
    { label: 'How do I train the AI?', answer: 'You can train Noant in seconds by uploading a CSV spreadsheet containing Q&A pairs, creating manual training categories, or typing custom answers directly in the Noant dashboard.' }
  ]

  const handleSendMessage = (text: string) => {
    if (!text.trim() || isTyping) return

    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      sender: 'user',
      text,
      timestamp: 'Just now'
    }
    setMessages((prev) => [...prev, userMsg])
    setInputValue('')
    setIsTyping(true)

    // Simulate AI response
    setTimeout(() => {
      let aiText = "That's a great question! I'm trained on your specific business documentation to answer that instantly. To connect this live AI assistant to your channels, sign up for a free trial."
      
      const matchedPreset = presetQuestions.find(q => q.label.toLowerCase() === text.toLowerCase())
      if (matchedPreset) {
        aiText = matchedPreset.answer
      } else if (text.toLowerCase().includes('price') || text.toLowerCase().includes('cost')) {
        aiText = 'Noant plans start at $0/mo for our Free tier (1 AI Agent, 100 chats/mo). Our Pro plan is $49/mo (unlimited chats, custom WhatsApp/Telegram sync) and Enterprise starts at $149/mo.'
      } else if (text.toLowerCase().includes('demo') || text.toLowerCase().includes('try')) {
        aiText = 'You are playing with it right now! Noant runs a similar custom-trained model on your website widget. Create a free account to customize your bot persona!'
      }

      const aiMsg: ChatMessage = {
        id: (Date.now() + 1).toString(),
        sender: 'ai',
        text: aiText,
        timestamp: 'Just now'
      }
      setMessages((prev) => [...prev, aiMsg])
      setIsTyping(false)
    }, 1500)
  }

  // Smooth scroll helper
  const scrollToId = (id: string) => {
    const el = document.getElementById(id)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth' })
      setMobileMenuOpen(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#000000] text-[#e5e5e5] font-sans antialiased overflow-x-hidden">
      {/* Decorative Background Gradients */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-[600px] overflow-hidden pointer-events-none z-0">
        <div className="absolute top-[-20%] left-[20%] w-[500px] h-[500px] rounded-full bg-sky-500/10 blur-[120px] animate-pulse duration-[8s]" />
        <div className="absolute top-[-10%] right-[20%] w-[450px] h-[450px] rounded-full bg-indigo-500/10 blur-[130px] animate-pulse duration-[12s]" />
      </div>

      {/* Floating Navbar */}
      <header className="fixed top-0 left-0 right-0 z-50 bg-[#0d0d0d]/80 backdrop-blur-md border-b border-[#262626] transition-all duration-300">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          {/* Logo */}
          <div className="flex items-center gap-2.5 cursor-pointer" onClick={() => scrollToId('hero')}>
            <svg className="w-7 h-7 text-sky-400" viewBox="0 0 200 200" fill="none">
              <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" />
              <circle cx="100" cy="100" r="70" fill="currentColor" />
              <circle cx="80" cy="100" r="6" fill="#000000" />
              <circle cx="100" cy="100" r="8" fill="#000000" />
              <circle cx="120" cy="100" r="10" fill="#000000" />
            </svg>
            <span className="text-lg font-bold tracking-widest lowercase text-sky-400">noant</span>
          </div>

          {/* Desktop Nav Links */}
          <nav className="hidden md:flex items-center gap-8">
            <button onClick={() => scrollToId('features')} className="text-sm text-zinc-400 hover:text-white transition-colors">Features</button>
            <button onClick={() => scrollToId('channels')} className="text-sm text-zinc-400 hover:text-white transition-colors">Channels</button>
            <button onClick={() => scrollToId('playground')} className="text-sm text-zinc-400 hover:text-white transition-colors">Interactive Demo</button>
            <button onClick={() => scrollToId('pricing')} className="text-sm text-zinc-400 hover:text-white transition-colors">Pricing</button>
          </nav>

          {/* Auth Action Buttons */}
          <div className="hidden md:flex items-center gap-4">
            {user ? (
              <Link
                to="/chats"
                className="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-semibold bg-sky-500 hover:bg-sky-600 text-white shadow-lg shadow-sky-500/20 active:scale-[0.98] transition-all duration-200"
              >
                Go to Dashboard
                <ArrowRight className="w-4 h-4" />
              </Link>
            ) : (
              <>
                <Link to="/login" className="text-sm font-semibold text-zinc-300 hover:text-white transition-colors px-3 py-2">
                  Sign in
                </Link>
                <Link
                  to="/signup"
                  className="inline-flex items-center gap-1 px-4 py-2 rounded-xl text-sm font-semibold bg-zinc-900 hover:bg-zinc-800 text-white border border-zinc-800 hover:border-zinc-700 active:scale-[0.98] transition-all duration-200"
                >
                  Get started
                </Link>
              </>
            )}
          </div>

          {/* Mobile Menu Toggle */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="md:hidden p-2 text-zinc-400 hover:text-white focus:outline-none"
          >
            {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {/* Mobile Dropdown Nav */}
        {mobileMenuOpen && (
          <div className="md:hidden bg-[#0d0d0d] border-b border-[#262626] px-6 py-4 space-y-4 animate-fade-in">
            <button onClick={() => scrollToId('features')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Features</button>
            <button onClick={() => scrollToId('channels')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Channels</button>
            <button onClick={() => scrollToId('playground')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Interactive Demo</button>
            <button onClick={() => scrollToId('pricing')} className="block w-full text-left py-2 text-sm text-zinc-400 hover:text-white transition-colors">Pricing</button>
            <div className="pt-4 border-t border-[#262626] flex flex-col gap-3">
              {user ? (
                <Link
                  to="/chats"
                  className="w-full text-center py-2.5 rounded-xl text-sm font-semibold bg-sky-500 text-white"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  Go to Dashboard
                </Link>
              ) : (
                <>
                  <Link
                    to="/login"
                    className="w-full text-center py-2.5 rounded-xl text-sm font-semibold text-zinc-300 hover:text-white transition-colors border border-zinc-850"
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    Sign in
                  </Link>
                  <Link
                    to="/signup"
                    className="w-full text-center py-2.5 rounded-xl text-sm font-semibold bg-zinc-900 text-white border border-zinc-800"
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    Get started
                  </Link>
                </>
              )}
            </div>
          </div>
        )}
      </header>

      {/* Hero Section */}
      <section id="hero" className="relative pt-32 pb-24 md:pt-40 md:pb-32 px-6 overflow-hidden">
        <div className="max-w-4xl mx-auto text-center space-y-8 relative z-10">
          {/* Badge */}
          <div className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 text-xs font-semibold tracking-wider uppercase animate-fade-in">
            <Sparkles className="w-3.5 h-3.5" />
            <span>Omnichannel Support Platform</span>
          </div>

          {/* Heading */}
          <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold text-white tracking-tight leading-[1.1] animate-slide-up">
            Customer Support, <br className="hidden sm:inline" />
            <span className="bg-gradient-to-r from-sky-400 via-sky-500 to-indigo-500 bg-clip-text text-transparent">
              Autopilot & Human
            </span>{' '}
            in Sync
          </h1>

          {/* Subtitle */}
          <p className="text-base sm:text-lg md:text-xl text-zinc-400 max-w-2xl mx-auto leading-relaxed font-light">
            Empower your team with a custom-trained AI assistant that resolves queries on WhatsApp, Telegram, and your website widget instantly. When confidence drops, support agents take over seamlessly.
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
            <Link
              to={user ? "/chats" : "/signup"}
              className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-3.5 rounded-xl text-sm font-semibold bg-sky-500 hover:bg-sky-600 text-white shadow-lg shadow-sky-500/20 hover:shadow-sky-500/30 active:scale-[0.98] transition-all duration-200 group"
            >
              {user ? 'Go to Dashboard' : 'Start Free Trial'}
              <ArrowRight className="w-4 h-4 group-hover:translate-x-0.5 transition-transform" />
            </Link>
            <button
              onClick={() => scrollToId('playground')}
              className="w-full sm:w-auto inline-flex items-center justify-center gap-1.5 px-6 py-3.5 rounded-xl text-sm font-semibold bg-zinc-900 hover:bg-zinc-800 text-white border border-zinc-800 hover:border-zinc-700 active:scale-[0.98] transition-all duration-200"
            >
              Test AI Playground
            </button>
          </div>
        </div>

        {/* Dashboard Mockup Showcase */}
        <div className="max-w-5xl mx-auto mt-16 md:mt-24 px-4 sm:px-6 relative z-10 animate-slide-up">
          {/* Subtle glow border */}
          <div className="absolute inset-0 bg-gradient-to-tr from-sky-500/10 via-indigo-500/10 to-transparent rounded-2xl blur-xl opacity-60 pointer-events-none" />
          
          <div className="relative bg-[#0d0d0d] border border-[#262626] rounded-2xl overflow-hidden shadow-2xl">
            {/* Window bar */}
            <div className="h-11 bg-[#0d0d0d] border-b border-[#262626] px-4 flex items-center justify-between">
              <div className="flex gap-2">
                <span className="w-3 h-3 rounded-full bg-red-500/30" />
                <span className="w-3 h-3 rounded-full bg-yellow-500/30" />
                <span className="w-3 h-3 rounded-full bg-green-500/30" />
              </div>
              <span className="text-[11px] text-zinc-500 font-medium tracking-wide">noant-dashboard.local</span>
              <div className="w-10" />
            </div>

            {/* Mock Layout */}
            <div className="grid grid-cols-12 h-[340px] sm:h-[420px] md:h-[480px] bg-[#000000]">
              {/* Mock Sidebar */}
              <div className="hidden sm:block col-span-3 border-r border-[#1a1a1a] p-3 space-y-4">
                <div className="h-8 bg-zinc-900 rounded-lg animate-pulse" />
                <div className="space-y-2.5">
                  <div className="h-7 bg-sky-950/20 text-sky-400 rounded-lg flex items-center px-2 text-xs font-semibold gap-2 border border-sky-950">
                    <LayoutGrid className="w-3.5 h-3.5" /> Overview
                  </div>
                  <div className="h-7 hover:bg-zinc-950 text-zinc-500 rounded-lg flex items-center px-2 text-xs gap-2">
                    <MessageSquare className="w-3.5 h-3.5" /> Conversations
                  </div>
                  <div className="h-7 hover:bg-zinc-950 text-zinc-500 rounded-lg flex items-center px-2 text-xs gap-2">
                    <GraduationCap className="w-3.5 h-3.5" /> Teach AI
                  </div>
                  <div className="h-7 hover:bg-zinc-950 text-zinc-500 rounded-lg flex items-center px-2 text-xs gap-2">
                    <Link2 className="w-3.5 h-3.5" /> Integrations
                  </div>
                </div>
              </div>

              {/* Mock Main Content */}
              <div className="col-span-12 sm:col-span-9 p-4 md:p-6 space-y-6 overflow-y-auto">
                <div className="flex justify-between items-center">
                  <div className="space-y-1">
                    <h3 className="text-sm font-bold text-white">Live Operations Overview</h3>
                    <p className="text-[10px] text-zinc-500">Real-time resolution rates and connected channel status.</p>
                  </div>
                  <span className="inline-flex items-center gap-1 text-[9px] text-green-400 bg-green-950/30 border border-green-900/50 px-2 py-0.5 rounded-full font-semibold">
                    <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-ping" /> Live Syncing
                  </span>
                </div>

                {/* Dashboard Stats */}
                <div className="grid grid-cols-3 gap-3 md:gap-4">
                  <div className="bg-[#0d0d0d] border border-[#1a1a1a] rounded-xl p-3.5 space-y-1">
                    <span className="text-[9px] text-zinc-500 uppercase font-semibold">AI Resolution</span>
                    <div className="text-base sm:text-2xl font-extrabold text-sky-400">84.2%</div>
                  </div>
                  <div className="bg-[#0d0d0d] border border-[#1a1a1a] rounded-xl p-3.5 space-y-1">
                    <span className="text-[9px] text-zinc-500 uppercase font-semibold">Open Chats</span>
                    <div className="text-base sm:text-2xl font-extrabold text-white">12</div>
                  </div>
                  <div className="bg-[#0d0d0d] border border-[#1a1a1a] rounded-xl p-3.5 space-y-1">
                    <span className="text-[9px] text-zinc-500 uppercase font-semibold">Handoffs</span>
                    <div className="text-base sm:text-2xl font-extrabold text-indigo-400">3</div>
                  </div>
                </div>

                {/* Simulated Chat Feed inside Mockup */}
                <div className="border border-[#1a1a1a] rounded-xl p-4 bg-[#070707] space-y-3">
                  <div className="flex justify-between items-center text-[10px] text-zinc-500 border-b border-[#1a1a1a] pb-2">
                    <span className="font-semibold text-zinc-400">WhatsApp Session — Active Chat</span>
                    <span>+44 7700 900123</span>
                  </div>
                  <div className="space-y-2.5">
                    <div className="flex gap-2">
                      <div className="w-5 h-5 rounded-full bg-zinc-800 text-white flex items-center justify-center text-[9px] font-bold">C</div>
                      <div className="bg-[#111111] text-zinc-300 text-[11px] p-2.5 rounded-r-xl rounded-bl-xl max-w-[80%] border border-[#1e1e2e]">
                        Hi, do you offer weekend shipping?
                      </div>
                    </div>
                    <div className="flex gap-2 justify-end">
                      <div className="bg-sky-950/40 text-sky-300 text-[11px] p-2.5 rounded-l-xl rounded-br-xl max-w-[80%] border border-sky-900/50">
                        Hello! Yes, shipping is processed 7 days a week. Orders placed before 2 PM on weekends are shipped the same day. 🚚
                      </div>
                      <div className="w-5 h-5 rounded-full bg-sky-500 text-white flex items-center justify-center text-[9px] font-bold">A</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Feature Showcase Grid */}
      <section id="features" className="py-24 px-6 border-t border-[#1a1a1a] bg-[#030303]">
        <div className="max-w-7xl mx-auto space-y-16">
          <div className="max-w-2xl mx-auto text-center space-y-4">
            <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
              SaaS Automation Meets Human Control
            </h2>
            <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
              Noant bridges the gap between hyper-fast autonomous AI responses and the high-fidelity touch of support agents.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 sm:gap-8">
            {/* Feature 1 */}
            <div className="bg-[#0d0d0d] border border-[#262626] hover:border-sky-500/50 rounded-2xl p-6 transition-all duration-300 group">
              <div className="w-10 h-10 rounded-xl bg-sky-500/10 flex items-center justify-center text-sky-400 mb-6 group-hover:scale-110 transition-transform">
                <Bot className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-2">Autonomous AI Autopilot</h3>
              <p className="text-sm text-zinc-400 leading-relaxed">
                Automatically resolves common queries, checks product inventory, and coordinates details with customers instantly using custom-trained models.
              </p>
            </div>

            {/* Feature 2 */}
            <div className="bg-[#0d0d0d] border border-[#262626] hover:border-indigo-500/50 rounded-2xl p-6 transition-all duration-300 group">
              <div className="w-10 h-10 rounded-xl bg-indigo-500/10 flex items-center justify-center text-indigo-400 mb-6 group-hover:scale-110 transition-transform">
                <Sparkles className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-2">Human-in-the-Loop Handoff</h3>
              <p className="text-sm text-zinc-400 leading-relaxed">
                When the AI encounters ambiguous questions or low similarity matches, the conversation escalates instantly so support agents can take over.
              </p>
            </div>

            {/* Feature 3 */}
            <div className="bg-[#0d0d0d] border border-[#262626] hover:border-emerald-500/50 rounded-2xl p-6 transition-all duration-300 group">
              <div className="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center text-emerald-400 mb-6 group-hover:scale-110 transition-transform">
                <GraduationCap className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-2">No-Hallucination Training</h3>
              <p className="text-sm text-zinc-400 leading-relaxed">
                Upload CSV sheets or add training categories. Noant AI only answers questions matched with high similarity, guaranteeing accurate answers.
              </p>
            </div>

            {/* Feature 4 */}
            <div className="bg-[#0d0d0d] border border-[#262626] hover:border-purple-500/50 rounded-2xl p-6 transition-all duration-300 group">
              <div className="w-10 h-10 rounded-xl bg-purple-500/10 flex items-center justify-center text-purple-400 mb-6 group-hover:scale-110 transition-transform">
                <Zap className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-2">Sub-10ms WebSockets</h3>
              <p className="text-sm text-zinc-400 leading-relaxed">
                Real-time connection syncing, typing states, and status badges are broadcasted concurrently with thread safety using our custom sync engine.
              </p>
            </div>

            {/* Feature 5 */}
            <div className="bg-[#0d0d0d] border border-[#262626] hover:border-amber-500/50 rounded-2xl p-6 transition-all duration-300 group">
              <div className="w-10 h-10 rounded-xl bg-amber-500/10 flex items-center justify-center text-amber-400 mb-6 group-hover:scale-110 transition-transform">
                <Shield className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-2">Robust Offline Resiliency</h3>
              <p className="text-sm text-zinc-400 leading-relaxed">
                A custom hook tracks connection status dynamically, locks crucial input components during disconnects, and restores seamlessly on connection.
              </p>
            </div>

            {/* Feature 6 */}
            <div className="bg-[#0d0d0d] border border-[#262626] hover:border-rose-500/50 rounded-2xl p-6 transition-all duration-300 group">
              <div className="w-10 h-10 rounded-xl bg-rose-500/10 flex items-center justify-center text-rose-400 mb-6 group-hover:scale-110 transition-transform">
                <BarChart3 className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-2">Lead Tracking & Insights</h3>
              <p className="text-sm text-zinc-400 leading-relaxed">
                Auto-classifies high-intent conversations, tags user contact profiles, and captures pipeline leads directly into your dashboard.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Connected Channels Showcase */}
      <section id="channels" className="py-24 px-6 border-t border-[#1a1a1a] relative overflow-hidden">
        {/* Glow */}
        <div className="absolute top-[30%] right-[-10%] w-[300px] h-[300px] rounded-full bg-indigo-500/5 blur-[90px] pointer-events-none" />

        <div className="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
          <div className="lg:col-span-5 space-y-6">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-[10px] font-bold uppercase tracking-wider">
              <Globe className="w-3.5 h-3.5" />
              <span>Omnichannel Support</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-extrabold text-white leading-tight">
              Connect Your Support Wherever Customers Are
            </h2>
            <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
              Noant centralizes messages from multiple channels into a single inbox, allowing your agents to manage multiple platforms concurrently.
            </p>

            <ul className="space-y-3.5 pt-2">
              <li className="flex items-start gap-3">
                <span className="w-5 h-5 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center shrink-0 mt-0.5"><Check className="w-3.5 h-3.5" /></span>
                <span className="text-zinc-300 text-sm">Self-hosted OpenWA WhatsApp container with cookie logouts.</span>
              </li>
              <li className="flex items-start gap-3">
                <span className="w-5 h-5 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center shrink-0 mt-0.5"><Check className="w-3.5 h-3.5" /></span>
                <span className="text-zinc-300 text-sm">Telegram Bot integration with webhook-driven chat handlers.</span>
              </li>
              <li className="flex items-start gap-3">
                <span className="w-5 h-5 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center shrink-0 mt-0.5"><Check className="w-3.5 h-3.5" /></span>
                <span className="text-zinc-300 text-sm">Embeddable Web Widget with dynamic configuration and styling.</span>
              </li>
            </ul>
          </div>

          <div className="lg:col-span-7 grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/* WhatsApp Integration card */}
            <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-6 flex flex-col justify-between h-[200px] hover:-translate-y-1 duration-350">
              <div className="flex justify-between items-start">
                <div className="w-10 h-10 rounded-xl bg-green-500/10 text-green-400 flex items-center justify-center"><Link2 className="w-5 h-5" /></div>
                <span className="text-[10px] text-green-400 bg-green-950/30 border border-green-900/50 px-2 py-0.5 rounded-full font-semibold">Active</span>
              </div>
              <div>
                <h3 className="font-bold text-white text-base mb-1">WhatsApp Channel</h3>
                <p className="text-xs text-zinc-400 leading-relaxed">Fully synchronized phone pairing, contact avatars, display names, and direct check validations.</p>
              </div>
            </div>

            {/* Telegram Integration card */}
            <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-6 flex flex-col justify-between h-[200px] hover:-translate-y-1 duration-350">
              <div className="flex justify-between items-start">
                <div className="w-10 h-10 rounded-xl bg-sky-500/10 text-sky-400 flex items-center justify-center"><Globe className="w-5 h-5" /></div>
                <span className="text-[10px] text-sky-400 bg-sky-950/30 border border-sky-900/50 px-2 py-0.5 rounded-full font-semibold">Configured</span>
              </div>
              <div>
                <h3 className="font-bold text-white text-base mb-1">Telegram Bot</h3>
                <p className="text-xs text-zinc-400 leading-relaxed">Connects to your bot token, processes message queues, and broadcasts custom AI responses directly.</p>
              </div>
            </div>

            {/* Web Widget card */}
            <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-6 flex flex-col justify-between h-[200px] hover:-translate-y-1 duration-350">
              <div className="flex justify-between items-start">
                <div className="w-10 h-10 rounded-xl bg-purple-500/10 text-purple-400 flex items-center justify-center"><Code2 className="w-5 h-5" /></div>
                <span className="text-[10px] text-purple-400 bg-purple-950/30 border border-purple-900/50 px-2 py-0.5 rounded-full font-semibold">Embeddable</span>
              </div>
              <div>
                <h3 className="font-bold text-white text-base mb-1">Web Chat Widget</h3>
                <p className="text-xs text-zinc-400 leading-relaxed">Embed a highly stylized glassmorphism chat widget onto your website with single line JS script.</p>
              </div>
            </div>

            {/* Facebook/Instagram card */}
            <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-6 flex flex-col justify-between h-[200px] hover:-translate-y-1 duration-350">
              <div className="flex justify-between items-start">
                <div className="w-10 h-10 rounded-xl bg-zinc-900 text-zinc-500 flex items-center justify-center"><Link2 className="w-5 h-5" /></div>
                <span className="text-[10px] text-zinc-500 bg-zinc-950 border border-zinc-800 px-2 py-0.5 rounded-full font-semibold">Coming Soon</span>
              </div>
              <div>
                <h3 className="font-bold text-white text-base mb-1">Social Integrations</h3>
                <p className="text-xs text-zinc-400 leading-relaxed">Direct messaging synchronization for Facebook Page and Instagram Business accounts is under development.</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Interactive AI Playground */}
      <section id="playground" className="py-24 px-6 border-t border-[#1a1a1a] bg-[#030303] relative">
        <div className="absolute top-[10%] left-[10%] w-[250px] h-[250px] rounded-full bg-sky-500/5 blur-[80px] pointer-events-none" />

        <div className="max-w-4xl mx-auto space-y-12">
          <div className="text-center space-y-4 max-w-2xl mx-auto">
            <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
              Try the Interactive AI Playground
            </h2>
            <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
              Test Noant's AI in real-time. Choose a preset training question below, or type your own question to see how Noant responds from its custom knowledge base.
            </p>
          </div>

          <div className="relative max-w-2xl mx-auto bg-[#0d0d0d] border border-[#262626] rounded-2xl overflow-hidden shadow-2xl">
            {/* Widget Header */}
            <div className="bg-[#0f0f0f] border-b border-[#262626] px-5 py-4 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center">
                  <Bot className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="text-xs font-bold text-white flex items-center gap-1.5">
                    Noant Assistant
                    <span className="w-1.5 h-1.5 rounded-full bg-green-400 animate-ping" />
                  </h3>
                  <p className="text-[10px] text-zinc-500">Autonomous Assistant Demo</p>
                </div>
              </div>
            </div>

            {/* Chat Body */}
            <div className="h-[280px] sm:h-[320px] p-5 overflow-y-auto space-y-4">
              {messages.map((msg) => (
                <div key={msg.id} className={`flex gap-3 ${msg.sender === 'user' ? 'justify-end' : ''}`}>
                  {msg.sender === 'ai' && (
                    <div className="w-6 h-6 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">
                      N
                    </div>
                  )}
                  <div className={`p-3 rounded-2xl text-[12px] leading-relaxed max-w-[80%] border ${
                    msg.sender === 'user'
                      ? 'bg-sky-500 text-white border-sky-400 rounded-tr-none'
                      : 'bg-[#141414] text-zinc-300 border-[#262626] rounded-tl-none'
                  }`}>
                    {msg.text}
                  </div>
                </div>
              ))}

              {isTyping && (
                <div className="flex gap-3">
                  <div className="w-6 h-6 rounded-full bg-sky-500/10 text-sky-400 flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">
                    N
                  </div>
                  <div className="bg-[#141414] text-zinc-400 text-[12px] p-3 rounded-2xl rounded-tl-none border border-[#262626] flex items-center gap-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-zinc-600 animate-bounce" style={{ animationDelay: '0ms' }} />
                    <span className="w-1.5 h-1.5 rounded-full bg-zinc-600 animate-bounce" style={{ animationDelay: '150ms' }} />
                    <span className="w-1.5 h-1.5 rounded-full bg-zinc-600 animate-bounce" style={{ animationDelay: '300ms' }} />
                  </div>
                </div>
              )}
            </div>

            {/* Preset Training Questions */}
            <div className="px-5 py-3 bg-[#0a0a0a] border-t border-[#1d1d1d] flex flex-wrap gap-2">
              <span className="text-[10px] text-zinc-500 flex items-center w-full mb-1">Click a training query to test:</span>
              {presetQuestions.map((q) => (
                <button
                  key={q.label}
                  onClick={() => handleSendMessage(q.label)}
                  disabled={isTyping}
                  className="text-[11px] bg-zinc-900 hover:bg-zinc-800 text-zinc-300 hover:text-white border border-zinc-800 px-3 py-1.5 rounded-full transition-all duration-200 text-left active:scale-[0.98] disabled:opacity-50"
                >
                  {q.label}
                </button>
              ))}
            </div>

            {/* Chat Input */}
            <form
              onSubmit={(e) => {
                e.preventDefault()
                handleSendMessage(inputValue)
              }}
              className="bg-[#0f0f0f] border-t border-[#262626] p-4 flex gap-3"
            >
              <input
                type="text"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                placeholder="Ask about pricing, manual takeover, or weekend shipping..."
                disabled={isTyping}
                className="flex-1 bg-[#141414] border border-[#262626] rounded-xl px-4 py-2.5 text-xs text-zinc-350 placeholder-zinc-600 focus:outline-none focus:border-sky-500 transition-colors disabled:opacity-50"
              />
              <button
                type="submit"
                disabled={!inputValue.trim() || isTyping}
                className="px-4 py-2.5 rounded-xl text-xs font-semibold bg-sky-500 hover:bg-sky-600 text-white active:scale-[0.98] transition-all disabled:opacity-50"
              >
                Send
              </button>
            </form>
          </div>
        </div>
      </section>

      {/* Pricing Section */}
      <section id="pricing" className="py-24 px-6 border-t border-[#1a1a1a]">
        <div className="max-w-7xl mx-auto space-y-16">
          <div className="text-center space-y-4 max-w-2xl mx-auto">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-sky-500/10 border border-sky-500/20 text-sky-400 text-[10px] font-bold uppercase tracking-wider">
              <Zap className="w-3.5 h-3.5" />
              <span>Simple, transparent pricing</span>
            </div>
            <h2 className="text-2xl sm:text-3xl font-extrabold text-white">
              Plans that scale with your growth
            </h2>
            <p className="text-zinc-400 text-sm sm:text-base leading-relaxed">
              Start for free and upgrade as your team grows and takes on more conversations.
            </p>

            {/* Toggle Billing */}
            <div className="flex items-center justify-center gap-3 pt-4">
              <span className={`text-xs ${!isAnnual ? 'text-white font-semibold' : 'text-zinc-500'}`}>Monthly</span>
              <button
                onClick={() => setIsAnnual(!isAnnual)}
                className="w-10 h-6 rounded-full bg-zinc-800 p-1 flex items-center relative transition-colors"
              >
                <div className={`w-4 h-4 rounded-full bg-sky-400 transition-all duration-300 ${isAnnual ? 'translate-x-4' : 'translate-x-0'}`} />
              </button>
              <span className={`text-xs ${isAnnual ? 'text-white font-semibold' : 'text-zinc-500'} flex items-center gap-1.5`}>
                Annual (Save 20%)
                <span className="text-[10px] bg-green-500/10 border border-green-500/20 text-green-400 px-1.5 py-0.5 rounded-full font-bold">20% Off</span>
              </span>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto items-stretch">
            {/* Free tier */}
            <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-8 flex flex-col justify-between relative hover:border-zinc-700 duration-300">
              <div className="space-y-6">
                <div className="space-y-2">
                  <h3 className="text-lg font-bold text-white">Free Sandbox</h3>
                  <p className="text-xs text-zinc-500">Perfect for initial testing and local bot tuning.</p>
                </div>
                <div className="flex items-baseline gap-1 text-white">
                  <span className="text-3xl sm:text-4xl font-extrabold">$0</span>
                  <span className="text-xs text-zinc-500">/mo</span>
                </div>
                <ul className="space-y-3 pt-4 border-t border-[#1a1a1a]">
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 100 AI chats / month</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 1 Connected Web Widget</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Manual Q&A FAQ training</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400 text-zinc-600"><X className="w-4 h-4 text-zinc-700 shrink-0" /> No WhatsApp integration</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400 text-zinc-600"><X className="w-4 h-4 text-zinc-700 shrink-0" /> No Telegram integration</li>
                </ul>
              </div>
              <Link to="/signup" className="mt-8 w-full text-center py-2.5 rounded-xl text-xs font-semibold bg-zinc-900 border border-zinc-800 text-white hover:bg-zinc-800 transition-colors">
                Get started for free
              </Link>
            </div>

            {/* Pro tier */}
            <div className="bg-[#0d0d0d] border-2 border-sky-500 rounded-2xl p-8 flex flex-col justify-between relative hover:border-sky-400 duration-300">
              <span className="absolute top-0 right-6 -translate-y-1/2 bg-sky-500 text-white text-[9px] font-bold uppercase tracking-wider px-2.5 py-1 rounded-full">Most Popular</span>
              <div className="space-y-6">
                <div className="space-y-2">
                  <h3 className="text-lg font-bold text-white">Noant Pro</h3>
                  <p className="text-xs text-zinc-500">For active businesses running omnichannel customer support.</p>
                </div>
                <div className="flex items-baseline gap-1 text-white">
                  <span className="text-3xl sm:text-4xl font-extrabold">${isAnnual ? '39' : '49'}</span>
                  <span className="text-xs text-zinc-500">/mo</span>
                </div>
                <ul className="space-y-3 pt-4 border-t border-[#1a1a1a]">
                  <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Unlimited AI chats</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 1 WhatsApp Session (OpenWA)</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> 1 Telegram Support Bot</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Spreadsheet CSV training imports</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-300"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Up to 3 team members</li>
                </ul>
              </div>
              <Link to="/signup" className="mt-8 w-full text-center py-2.5 rounded-xl text-xs font-semibold bg-sky-500 text-white hover:bg-sky-600 shadow-lg shadow-sky-500/20 transition-colors">
                Start Pro Trial
              </Link>
            </div>

            {/* Enterprise tier */}
            <div className="bg-[#0d0d0d] border border-[#262626] rounded-2xl p-8 flex flex-col justify-between relative hover:border-zinc-700 duration-300">
              <div className="space-y-6">
                <div className="space-y-2">
                  <h3 className="text-lg font-bold text-white">Noant Enterprise</h3>
                  <p className="text-xs text-zinc-500">For organizations requiring tailored scaling and custom models.</p>
                </div>
                <div className="flex items-baseline gap-1 text-white">
                  <span className="text-3xl sm:text-4xl font-extrabold">${isAnnual ? '119' : '149'}</span>
                  <span className="text-xs text-zinc-500">/mo</span>
                </div>
                <ul className="space-y-3 pt-4 border-t border-[#1a1a1a]">
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Multiple WhatsApp & Telegram lines</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Unlimited team members</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Custom LLM intent models (Groq)</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> Dedicated priority support agent</li>
                  <li className="flex items-center gap-2.5 text-xs text-zinc-400"><Check className="w-4 h-4 text-sky-400 shrink-0" /> API integration endpoints</li>
                </ul>
              </div>
              <Link to="/signup" className="mt-8 w-full text-center py-2.5 rounded-xl text-xs font-semibold bg-zinc-900 border border-zinc-800 text-white hover:bg-zinc-800 transition-colors">
                Contact sales
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Footer Section */}
      <footer className="bg-[#070707] border-t border-[#1a1a1a] py-12 px-6 text-center text-xs text-zinc-650">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="flex items-center gap-2.5 text-sky-400">
            <svg className="w-5 h-5 shrink-0" viewBox="0 0 200 200" fill="none">
              <circle cx="100" cy="100" r="92" stroke="currentColor" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" />
              <circle cx="100" cy="100" r="70" fill="currentColor" />
            </svg>
            <span className="text-sm font-bold tracking-widest lowercase">noant</span>
          </div>
          <p>© {new Date().getFullYear()} Noant Customer Support. All rights reserved.</p>
          <div className="flex gap-4">
            <a href="#" className="hover:text-white transition-colors">Terms</a>
            <a href="#" className="hover:text-white transition-colors">Privacy</a>
            <a href="#" className="hover:text-white transition-colors">Documentation</a>
          </div>
        </div>
      </footer>
    </div>
  )
}
