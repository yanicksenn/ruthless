import React, { useState, useEffect } from 'react';
import { sessionClient, createOptions } from '../api/client';
import { Session } from '../api/ruthless';
import { useAuth } from '../context/AuthContext';
import { Plus, Play, Users, Hash, LogOut } from 'lucide-react';

interface LobbyProps {
  onJoinSession: (sessionId: string) => void;
}

export const Lobby: React.FC<LobbyProps> = ({ onJoinSession }) => {
  const { token, user, logout } = useAuth();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchSessions = async () => {
    try {
      const response = await sessionClient.listSessions({}, createOptions(token));
      setSessions(response.response.sessions);
    } catch (err) {
      console.error('Failed to list sessions:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSessions();
    const interval = setInterval(fetchSessions, 5000); // Poll every 5s
    return () => clearInterval(interval);
  }, [token]);

  const handleCreateSession = async () => {
    try {
      const response = await sessionClient.createSession({}, createOptions(token));
      onJoinSession(response.response.id);
    } catch (err) {
      console.error('Failed to create session:', err);
    }
  };

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="flex justify-between items-start mb-12">
        <div>
          <h1 className="text-5xl font-black tracking-tighter text-white">LOBBY</h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">Join a session or create your own</p>
          {user && (
            <div className="mt-4 flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xs font-bold ring-1 ring-primary/30">
                {user.name.slice(0, 2).toUpperCase()}
              </div>
              <span className="text-gray-300 font-bold text-sm tracking-tight">{user.name}</span>
            </div>
          )}
        </div>
        <div className="flex gap-3">
          <button
            onClick={handleCreateSession}
            className="bg-secondary hover:bg-secondary/80 text-black font-black px-6 py-3 rounded-2xl flex items-center gap-2 transition-all transform hover:scale-105 shadow-lg shadow-secondary/10"
          >
            <Plus size={20} />
            NEW SESSION
          </button>
          <button
            onClick={logout}
            className="p-3 bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white rounded-2xl transition-all ring-1 ring-white/5"
            title="Logout"
          >
            <LogOut size={20} />
          </button>
        </div>
      </header>

      {loading && sessions.length === 0 ? (
        <div className="flex justify-center p-12">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {sessions.map((session) => (
            <div
              key={session.id}
              className="glass p-6 rounded-2xl hover:border-primary/50 transition-all cursor-pointer group"
              onClick={() => onJoinSession(session.id)}
            >
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-2 text-primary">
                  <Hash size={16} />
                  <span className="font-mono text-sm">{session.id.substring(0, 8)}</span>
                </div>
                <div className="flex items-center gap-1.5 text-gray-400 bg-white/5 px-2.5 py-1 rounded-full text-xs font-bold">
                  <Users size={14} />
                  {session.playerIds.length} PLAYERS
                </div>
              </div>
              <h3 className="text-xl font-bold mb-6 group-hover:text-primary transition-colors">
                {session.ownerId.split('-')[0]}'s Session
              </h3>
              <div className="flex justify-between items-center">
                <div className="flex -space-x-2">
                   {session.playerIds.slice(0, 4).map((pid) => (
                     <div key={pid} className="w-8 h-8 rounded-full bg-surface border-2 border-background flex items-center justify-center text-[10px] font-bold">
                       {pid.slice(0, 2).toUpperCase()}
                     </div>
                   ))}
                   {session.playerIds.length > 4 && (
                     <div className="w-8 h-8 rounded-full bg-gray-800 border-2 border-background flex items-center justify-center text-[10px] font-bold">
                       +{session.playerIds.length - 4}
                     </div>
                   )}
                </div>
                <button className="text-xs font-black uppercase text-gray-400 group-hover:text-white transition-colors flex items-center gap-1">
                  JOIN ROOM <Play size={12} fill="currentColor" />
                </button>
              </div>
            </div>
          ))}
          {sessions.length === 0 && !loading && (
             <div className="md:col-span-2 glass p-12 rounded-3xl text-center border-dashed border-2 border-white/5">
                <p className="text-gray-500 font-bold italic">No active sessions. Be the first to start the chaos.</p>
             </div>
          )}
        </div>
      )}
    </div>
  );
};
