import React, { useState, useEffect } from 'react';
import { sessionClient, sessionInvitationClient, notificationClient, deckClient, createOptions } from '../api/client';
import { Session, SessionInvitation, SessionView, Deck, NotificationType } from '../api/ruthless';
import { useAuth } from '../context/AuthContext';
import { Plus, Play, Users, Hash, LogOut, Check, X } from 'lucide-react';
import { SessionCreationDialog } from './SessionCreationDialog';

type Tab = 'public' | 'active' | 'invitations';

interface LobbyProps {
  onJoinSession: (sessionId: string) => void;
}

export const Lobby: React.FC<LobbyProps> = ({ onJoinSession }) => {
  const { token, user, logout, limits } = useAuth();
  const [activeTab, setActiveTab] = useState<Tab>('public');
  
  const [publicSessions, setPublicSessions] = useState<Session[]>([]);
  const [activeSessions, setActiveSessions] = useState<Session[]>([]);
  const [invitations, setInvitations] = useState<SessionInvitation[]>([]);
  
  const [loading, setLoading] = useState(true);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [decks, setDecks] = useState<Deck[]>([]);
  const [hasNotifications, setHasNotifications] = useState(false);

  const checkNotifications = async () => {
    try {
      const res = await notificationClient.getNotifications({}, createOptions(token));
      const notifications = res.response.notifications || [];
      const hasPending = notifications.some(
        n => n.type === NotificationType.SESSION_INVITATIONS_PENDING && n.count > 0
      );
      setHasNotifications(hasPending);
    } catch (err) {
      console.error('Failed to get notifications:', err);
    }
  };

  const fetchPublicSessions = async () => {
    try {
      const response = await sessionClient.listSessions({ view: SessionView.PUBLIC_WAITING }, createOptions(token));
      setPublicSessions(response.response.sessions || []);
    } catch (err) {
      console.error('Failed to fetch public sessions:', err);
    }
  };

  const fetchActiveSessions = async () => {
    try {
      const response = await sessionClient.listSessions({ view: SessionView.ACTIVE }, createOptions(token));
      setActiveSessions(response.response.sessions || []);
    } catch (err) {
      console.error('Failed to fetch active sessions:', err);
    }
  };

  const fetchInvitations = async () => {
    try {
      const response = await sessionInvitationClient.listSessionInvitations({ pageSize: 50, pageNumber: 1 }, createOptions(token));
      setInvitations(response.response.invitations || []);
    } catch (err) {
      console.error('Failed to fetch session invitations:', err);
    }
  };

  const fetchData = async () => {
    await Promise.all([
      fetchPublicSessions(),
      fetchActiveSessions(),
      fetchInvitations()
    ]);
    if (loading) setLoading(false);
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000); // Poll every 5s
    return () => clearInterval(interval);
  }, [token]);

  useEffect(() => {
    checkNotifications();
    window.addEventListener('notifications-updated', checkNotifications);
    return () => {
      window.removeEventListener('notifications-updated', checkNotifications);
    };
  }, [token]);

  useEffect(() => {
    if (activeTab === 'invitations') {
      setHasNotifications(false);
      window.dispatchEvent(new Event('notifications-reset'));
      const reset = async () => {
        try {
          await notificationClient.resetNotificationCounter(
            { type: NotificationType.SESSION_INVITATIONS_PENDING },
            createOptions(token)
          );
        } catch (err) {
          console.error('Failed to reset notification counter:', err);
        }
      };
      reset();
    }
  }, [activeTab, token]);

  const handleOpenCreateModal = async () => {
    setLoading(true);
    try {
      const decksRes = await deckClient.listDecks({}, createOptions(token));
      setDecks(decksRes.response.decks || []);
      setIsCreateModalOpen(true);
    } catch (err) {
      console.error('Failed to fetch decks:', err);
      alert('Failed to fetch available decks');
    } finally {
      setLoading(false);
    }
  };

  const handleJoinSession = async (sessionId: string) => {
    try {
      await sessionClient.joinSession({ 
        sessionId,
        playerName: user?.name || ''
      }, createOptions(token));
      onJoinSession(sessionId);
    } catch (err) {
      console.error('Failed to join session:', err);
      alert('Failed to join session');
    }
  };

  const handleCreateSession = async (deckIds: string[], name: string) => {
    try {
      const response = await sessionClient.createSession({ deckIds, name }, createOptions(token));
      setIsCreateModalOpen(false);
      onJoinSession(response.response.id);
    } catch (err) {
      console.error('Failed to create session:', err);
      alert('Failed to create session');
    }
  };

  const handleRespondToInvitation = async (invitationId: string, accept: boolean) => {
    try {
      const res = await sessionInvitationClient.respondToSessionInvitation({ invitationId, accept }, createOptions(token));
      if (accept && res.response.sessionId) {
          onJoinSession(res.response.sessionId);
      } else {
          fetchData();
      }
    } catch (err: any) {
      alert(`Failed to respond to invitation: ${err.message || 'Unknown error'}`);
    }
  };

  const renderSessionCard = (session: Session) => (
    <div
      key={session.id}
      className="glass p-6 rounded-2xl transition-all cursor-pointer group hover:border-primary/50"
      onClick={() => handleJoinSession(session.id)}
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
      <h3 className="text-xl font-bold mb-6 transition-colors group-hover:text-primary text-white">
        {session.name}
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
        <div className="text-xs font-black uppercase transition-colors flex items-center gap-1 text-gray-400 group-hover:text-white">
          JOIN ROOM <Play size={12} fill="currentColor" />
        </div>
      </div>
    </div>
  );

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="flex justify-between items-start mb-8">
        <div>
          <h1 className="text-5xl font-black tracking-tighter text-white">SESSIONS</h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">Join a session or create your own</p>
          {user && (
            <div className="mt-4 flex items-center gap-2">
              <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-primary text-xs font-bold ring-1 ring-primary/30">
                {user.name.slice(0, 2).toUpperCase()}
              </div>
              <span className="text-gray-300 font-bold text-sm tracking-tight">
                {user.name}
                {user.identifier && <span className="text-gray-500 italic">#{user.identifier}</span>}
              </span>
            </div>
          )}
        </div>
        <div className="flex gap-3">
          <button
            onClick={handleOpenCreateModal}
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

      <div className="flex gap-4 mb-8 border-b border-white/10 pb-4">
        <button
          onClick={() => setActiveTab('public')}
          className={`px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase flex items-center gap-2 ${
            activeTab === 'public' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Public Sessions
        </button>
        <button
          onClick={() => setActiveTab('active')}
          className={`px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase flex items-center gap-2 ${
            activeTab === 'active' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Active Sessions
        </button>
        <button
          onClick={() => setActiveTab('invitations')}
          className={`relative px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase flex items-center gap-2 ${
            activeTab === 'invitations' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Invitations
          {hasNotifications && (
            <span className="w-2 h-2 bg-red-500 rounded-full inline-block" />
          )}
        </button>
      </div>

      <div className="glass p-8 rounded-[2.5rem] min-h-[500px] border border-white/5 shadow-2xl relative overflow-hidden">
        {loading && publicSessions.length === 0 && activeSessions.length === 0 && invitations.length === 0 && (
          <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-[2.5rem]">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
          </div>
        )}

        <div className="space-y-8">
          {activeTab === 'public' && (
            <>
              <div className="mb-4">
                <h2 className="text-3xl font-black text-white tracking-tight">Public Sessions</h2>
                <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">Available sessions waiting for players</p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {publicSessions.map(renderSessionCard)}
                {publicSessions.length === 0 && !loading && (
                   <div className="md:col-span-2 glass p-12 rounded-3xl text-center border-dashed border-2 border-white/5">
                      <p className="text-gray-500 font-bold italic">No public sessions waiting. Be the first to start the chaos.</p>
                   </div>
                )}
              </div>
            </>
          )}

          {activeTab === 'active' && (
            <>
              <div className="mb-4">
                <h2 className="text-3xl font-black text-white tracking-tight">Active Sessions</h2>
                <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">Sessions you are currently participating in</p>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {activeSessions.map(renderSessionCard)}
                {activeSessions.length === 0 && !loading && (
                   <div className="md:col-span-2 glass p-12 rounded-3xl text-center border-dashed border-2 border-white/5">
                      <p className="text-gray-500 font-bold italic">You are not participating in any active sessions.</p>
                   </div>
                )}
              </div>
            </>
          )}

          {activeTab === 'invitations' && (
            <>
              <div className="mb-4">
                <h2 className="text-3xl font-black text-white tracking-tight">Session Invitations</h2>
                <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">Friends inviting you to their games</p>
              </div>
              <div className="space-y-4">
                {invitations.length === 0 ? (
                  <div className="text-center py-20 text-gray-500 font-bold italic">
                    No pending session invitations.
                  </div>
                ) : (
                  invitations.map((inv) => (
                    <div key={inv.id} className="flex flex-col sm:flex-row items-center justify-between p-4 glass-light rounded-2xl border border-white/5 gap-4">
                      <div className="flex items-center gap-4 flex-1">
                        <div className="w-12 h-12 rounded-full bg-secondary/20 flex items-center justify-center text-secondary font-bold ring-1 ring-secondary/30">
                          {inv.sender?.name?.slice(0, 2).toUpperCase()}
                        </div>
                        <div>
                          <h3 className="text-white font-bold text-lg">{inv.sender?.name}</h3>
                          <p className="text-gray-500 text-sm font-bold tracking-tighter">Invited you to handle <span className="text-white">"{inv.session?.name}"</span></p>
                        </div>
                      </div>
                      <div className="flex gap-2 w-full sm:w-auto mt-4 sm:mt-0">
                        <button
                          onClick={() => handleRespondToInvitation(inv.id, true)}
                          className="flex-1 sm:flex-none p-3 px-6 bg-primary/20 hover:bg-primary text-primary hover:text-background rounded-xl transition-all font-bold flex items-center justify-center gap-2"
                        >
                          <Check size={18} /> ACCEPT
                        </button>
                        <button
                          onClick={() => handleRespondToInvitation(inv.id, false)}
                          className="flex-1 sm:flex-none p-3 px-6 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white rounded-xl transition-all font-bold flex items-center justify-center gap-2"
                        >
                          <X size={18} /> DECLINE
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </>
          )}
        </div>
      </div>

      <SessionCreationDialog
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onCreate={handleCreateSession}
        decks={decks}
        defaultName={user ? `${user.name}'s Session` : 'New Session'}
        limits={limits}
      />
    </div>
  );
};
