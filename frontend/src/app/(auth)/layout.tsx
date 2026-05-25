import { Outlet } from 'react-router-dom'

export default function AuthLayout() {
  return (
    <div className="h-screen flex overflow-hidden">
      {/* Left side - brand (desktop only) */}
      <div className="hidden lg:flex flex-1 bg-noant-black text-white flex-col justify-between p-12 relative overflow-hidden">
        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-16">
            {/* Exact logo - white version on dark bg */}
            <svg className="w-10 h-10" viewBox="0 0 200 200" fill="none">
              <circle cx="100" cy="100" r="92" stroke="white" strokeWidth="10" strokeDasharray="14 18" strokeLinecap="round" />
              <circle cx="100" cy="100" r="70" fill="white" />
              <circle cx="80" cy="100" r="6" fill="#0a0a0a" />
              <circle cx="100" cy="100" r="8" fill="#0a0a0a" />
              <circle cx="120" cy="100" r="10" fill="#0a0a0a" />
            </svg>
            <span className="text-2xl font-bold tracking-widest lowercase">noant</span>
          </div>
          <p className="text-3xl font-light leading-relaxed max-w-md">
            "I used to wake up to 47 unread messages. Now I wake up to closed sales."
          </p>
        </div>
        <div className="relative z-10 flex gap-12">
          <div>
            <div className="text-3xl font-bold text-noant-sky">500+</div>
            <div className="text-sm opacity-60 uppercase tracking-widest mt-1">Languages</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-noant-sky">24/7</div>
            <div className="text-sm opacity-60 uppercase tracking-widest mt-1">Always on</div>
          </div>
          <div>
            <div className="text-3xl font-bold text-noant-sky">&lt;2s</div>
            <div className="text-sm opacity-60 uppercase tracking-widest mt-1">Response time</div>
          </div>
        </div>
      </div>

      {/* Right side - form */}
      <div className="flex-1 flex items-center justify-center p-4 lg:p-12 bg-base overflow-y-auto lg:overflow-hidden">
        <div className="w-full max-w-sm">
          <Outlet />
        </div>
      </div>
    </div>
  )
}
