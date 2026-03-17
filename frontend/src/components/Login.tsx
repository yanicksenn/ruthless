import React from 'react';
import { useAuth } from '../context/AuthContext';
import { LogIn, UserPlus } from 'lucide-react';
import { GoogleLogin } from '@react-oauth/google';

export const Login: React.FC = () => {
  const { user, login, register, completeRegistration, isDevelopment } = useAuth();
  const [error, setError] = React.useState<string | null>(null);
  const [chosenName, setChosenName] = React.useState('');

  const handleLoginSuccess = async (credentialResponse: any) => {
    if (credentialResponse.credential) {
      setError(null);
      try {
        await login(credentialResponse.credential);
      } catch (err: any) {
        setError(err.message || 'Login failed. Are you registered?');
      }
    }
  };

  const handleRegisterSuccess = async (credentialResponse: any) => {
    if (credentialResponse.credential) {
      setError(null);
      try {
        await register(credentialResponse.credential);
      } catch (err: any) {
        setError(err.message || 'Registration failed.');
      }
    }
  };

  const handleCompleteRegistration = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!chosenName.trim()) return;

    setError(null);
    try {
      await completeRegistration(chosenName.trim());
    } catch (err: any) {
      setError(err.message || 'Registration completion failed.');
    }
  };

  if (user?.pendingCompletion) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <div className="glass p-10 rounded-3xl w-full max-w-md card-shadow border border-white/5 animate-in fade-in zoom-in duration-300">
          <div className="text-center mb-10">
            <div className="inline-block p-3 rounded-2xl bg-secondary/10 mb-4">
              <UserPlus size={32} className="text-secondary" />
            </div>
            <h1 className="text-3xl font-black tracking-tight text-white mb-2">PICK YOUR ALIAS</h1>
            <p className="text-gray-400 text-sm">Every terrible person needs a name.</p>
          </div>

          <form onSubmit={handleCompleteRegistration} className="space-y-6">
            <div className="space-y-2">
              <label className="text-xs font-bold text-gray-500 uppercase tracking-widest px-1">Chosen Name</label>
              <input
                autoFocus
                type="text"
                placeholder="e.g. TotalAsshat"
                value={chosenName}
                onChange={(e) => setChosenName(e.target.value)}
                className="w-full bg-white/5 border border-white/10 rounded-2xl px-6 py-4 text-white focus:outline-none focus:ring-2 focus:ring-secondary/50 transition-all font-bold text-lg"
              />
              <p className="text-[10px] text-gray-500 px-1 italic">
                A random 8-digit identifier will be appended to your name.
              </p>
            </div>

            {error && (
              <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 text-sm text-center">
                {error}
              </div>
            )}

            <div className="pt-2 flex gap-3">
              <button
                type="submit"
                disabled={!chosenName.trim()}
                className="w-full bg-secondary hover:bg-secondary/80 text-background py-4 rounded-2xl font-black transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-[0_0_20px_rgba(234,179,8,0.3)]"
              >
                COMPLETE REGISTRATION
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="glass p-10 rounded-3xl w-full max-w-md card-shadow border border-white/5">
        <div className="text-center mb-10">
          <div className="inline-block p-3 rounded-2xl bg-primary/10 mb-4">
            <LogIn size={32} className="text-primary" />
          </div>
          <h1 className="text-5xl font-black tracking-tighter text-white mb-2">RUTHLESS</h1>
          <p className="text-gray-400 uppercase text-xs tracking-[0.2em] font-bold">A Game for Terrible People</p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-500/10 border border-red-500/20 rounded-2xl text-red-400 text-sm text-center">
            {error}
          </div>
        )}
        
        <div className="space-y-6">
          <div className="space-y-3">
            <h2 className="text-xl font-bold text-white flex items-center gap-2">
              <LogIn size={18} className="text-primary" />
              Sign In
            </h2>
            <p className="text-sm text-gray-400 leading-relaxed">
              If you already have an account, sign in with Google to continue your streak of terror.
            </p>
            <div className="flex justify-center bg-white/5 p-4 rounded-2xl border border-white/5">
              <GoogleLogin
                onSuccess={handleLoginSuccess}
                onError={() => setError('Google Sign In failed')}
                theme="filled_black"
                shape="pill"
                text="signin_with"
                width="100%"
              />
            </div>
          </div>

          <div className="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />

          <div className="space-y-3">
            <h2 className="text-xl font-bold text-white flex items-center gap-2">
              <UserPlus size={18} className="text-secondary" />
              Register
            </h2>
            <p className="text-sm text-gray-400 leading-relaxed">
              New here? Register with Google to start offending your friends.
            </p>
            <div className="flex justify-center bg-white/5 p-4 rounded-2xl border border-white/5">
              <GoogleLogin
                onSuccess={handleRegisterSuccess}
                onError={() => setError('Google Registration failed')}
                theme="outline"
                shape="pill"
                text="signup_with"
                width="100%"
              />
            </div>
          </div>
          <div className="h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />

          {isDevelopment && (
            <div className="space-y-3">
              <h2 className="text-xl font-bold text-white flex items-center gap-2">
                <span className="p-1 rounded bg-gray-500/10"><UserPlus size={16} className="text-gray-400" /></span>
                Developer Access
              </h2>
              <p className="text-sm text-gray-400 leading-relaxed">
                Backend is in development mode. Use a developer token to login.
              </p>
              <form onSubmit={async (e) => {
                e.preventDefault();
                const target = e.target as typeof e.target & {
                  token: { value: string };
                };
                const tokenVal = target.token.value.trim();
                if (tokenVal) {
                  setError(null);
                  try {
                    await login(tokenVal);
                  } catch (err: any) {
                    // If login fails with PermissionDenied (code 7) or user not found, try register
                    if (err.code === "PERMISSION_DENIED" || err.message?.toLowerCase().includes("not registered") || err.message?.toLowerCase().includes("not found")) {
                      try {
                        await register(tokenVal);
                      } catch (regErr: any) {
                        setError(regErr.message || 'Developer registration failed.');
                      }
                    } else {
                      setError(err.message || 'Developer login failed.');
                    }
                  }
                }
              }} className="flex gap-2">
              <input
                name="token"
                type="text"
                placeholder="Name or Token"
                className="flex-1 bg-white/5 border border-white/10 rounded-xl px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
              />
              <button
                type="submit"
                className="bg-white/10 hover:bg-white/20 text-white px-4 py-2 rounded-xl text-sm font-bold transition-all"
              >
                GO
              </button>
            </form>
          </div>
        )}
      </div>

      <p className="mt-10 text-[10px] text-gray-600 text-center uppercase tracking-widest font-bold font-mono">
        Identity: {import.meta.env.VITE_GOOGLE_CLIENT_ID ? 'Configured' : 'Missing'} | Port: {window.location.port}
      </p>
    </div>
  </div>
);
};
