import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import { UserPlus, Users, Mail, Check, X, UserMinus, Plus, LogOut } from 'lucide-react';
import { friendClient, notificationClient, createOptions } from '../api/client';
import { Player, FriendInvitation, NotificationType } from '../api/ruthless';
import { CreationDialog } from './CreationDialog';

type Tab = 'friends' | 'invitations';

export const Friends: React.FC = () => {
  const { token, user, logout } = useAuth();
  const [activeTab, setActiveTab] = useState<Tab>('friends');
  const [friends, setFriends] = useState<Player[]>([]);
  const [invitations, setInvitations] = useState<FriendInvitation[]>([]);
  const [loading, setLoading] = useState(true);
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [hasNotifications, setHasNotifications] = useState(false);

  const [friendsPageNumber, setFriendsPageNumber] = useState(1);
  const [friendsTotalCount, setFriendsTotalCount] = useState(0);
  const [invitationsPageNumber, setInvitationsPageNumber] = useState(1);
  const [invitationsTotalCount, setInvitationsTotalCount] = useState(0);
  const pageSize = 10;

  const checkNotifications = async () => {
    try {
      const res = await notificationClient.getNotifications({}, createOptions(token));
      const notifications = res.response.notifications || [];
      const hasPending = notifications.some(
        n => n.type === NotificationType.FRIENDS_PENDING_INVITATIONS && n.count > 0
      );
      setHasNotifications(hasPending);
    } catch (err) {
      console.error('Failed to get notifications:', err);
    }
  };

  const fetchFriends = async () => {
    setLoading(true);
    try {
      const res = await friendClient.listFriends({
        pageSize,
        pageNumber: friendsPageNumber
      }, createOptions(token));
      setFriends(res.response.friends || []);
      setFriendsTotalCount(res.response.totalCount);
    } catch (err) {
      console.error('Failed to fetch friends:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchInvitations = async () => {
    setLoading(true);
    try {
      const res = await friendClient.listInvitations({
        pageSize,
        pageNumber: invitationsPageNumber
      }, createOptions(token));
      setInvitations(res.response.invitations || []);
      setInvitationsTotalCount(res.response.totalCount);
    } catch (err) {
      console.error('Failed to fetch invitations:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchData = async () => {
    await Promise.all([fetchFriends(), fetchInvitations()]);
  };

  useEffect(() => {
    if (activeTab === 'invitations') {
      setHasNotifications(false);
      window.dispatchEvent(new Event('notifications-reset'));
      
      const reset = async () => {
        try {
          await notificationClient.resetNotificationCounter(
            { type: NotificationType.FRIENDS_PENDING_INVITATIONS },
            createOptions(token)
          );
        } catch (err) {
          console.error('Failed to reset notification counter:', err);
        }
      };
      reset();
    }
  }, [activeTab, token]);

  useEffect(() => {
    fetchFriends();
  }, [token, friendsPageNumber]);

  useEffect(() => {
    fetchInvitations();
  }, [token, invitationsPageNumber]);

  useEffect(() => {
    checkNotifications();

    window.addEventListener('notifications-updated', checkNotifications);
    return () => {
      window.removeEventListener('notifications-updated', checkNotifications);
    };
  }, [token]);

  const handleInvite = async (identifier: string) => {
    try {
      await friendClient.inviteFriend({ identifier }, createOptions(token));
      alert('Invitation sent!');
      fetchFriends();
      fetchInvitations();
    } catch (err: any) {
      alert(`Failed to send invitation: ${err.message || 'Unknown error'}`);
    }
  };

  const handleRespond = async (invitationId: string, accept: boolean) => {
    try {
      await friendClient.respondToInvitation({ invitationId, accept }, createOptions(token));
      fetchFriends();
      fetchInvitations();
    } catch (err: any) {
      alert(`Failed to respond to invitation: ${err.message || 'Unknown error'}`);
    }
  };

  const handleRemoveFriend = async (friendId: string) => {
    if (!window.confirm('Are you sure you want to remove this friend?')) return;
    try {
      await friendClient.removeFriend({ friendId }, createOptions(token));
      fetchFriends();
    } catch (err: any) {
      alert(`Failed to remove friend: ${err.message || 'Unknown error'}`);
    }
  };

  return (
    <div className="max-w-6xl mx-auto p-4 py-12">
      <header className="flex justify-between items-start mb-8">
        <div>
          <h1 className="text-5xl font-black tracking-tighter text-white">FRIENDS</h1>
          <p className="text-gray-400 font-bold uppercase tracking-widest text-sm">
            Manage your social circle
          </p>
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
            onClick={() => setIsInviteDialogOpen(true)}
            className="bg-secondary hover:bg-secondary/80 text-black font-black px-6 py-3 rounded-2xl flex items-center gap-2 transition-all transform hover:scale-105 shadow-lg shadow-secondary/10"
          >
            <UserPlus size={20} />
            INVITE FRIEND
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
          onClick={() => setActiveTab('friends')}
          className={`px-6 py-2 rounded-xl font-black text-sm transition-all tracking-widest uppercase flex items-center gap-2 ${
            activeTab === 'friends' ? 'bg-primary text-background' : 'text-gray-400 hover:text-white hover:bg-white/5'
          }`}
        >
          Friends
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
        {loading && (
          <div className="absolute inset-0 bg-background/50 backdrop-blur-sm z-10 flex items-center justify-center rounded-[2.5rem]">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-primary"></div>
          </div>
        )}

        <div className="space-y-8">
          {activeTab === 'friends' ? (
            <>
              <div className="mb-4">
                <h2 className="text-3xl font-black text-white tracking-tight">Social Circle</h2>
                <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">Your collection of acquaintances</p>
              </div>

              <div className="space-y-4">
                {friends.length === 0 ? (
                  <div className="text-center py-20 text-gray-500 font-bold italic">
                    No friends yet. Start by inviting someone!
                  </div>
                ) : (
                  friends.map((friend) => (
                    <div key={friend.id} className="flex items-center justify-between p-4 glass-light rounded-2xl border border-white/5 group">
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-full bg-primary/20 flex items-center justify-center text-primary font-bold ring-1 ring-primary/30">
                          {friend.name.slice(0, 2).toUpperCase()}
                        </div>
                        <div>
                          <h3 className="text-white font-bold text-lg">{friend.name}</h3>
                          <p className="text-gray-500 text-sm font-mono tracking-tighter">#{friend.identifier}</p>
                        </div>
                      </div>
                      <button
                        onClick={() => handleRemoveFriend(friend.id)}
                        className="p-2 text-gray-500 hover:text-red-500 transition-colors opacity-0 group-hover:opacity-100"
                        title="Remove Friend"
                      >
                        <UserMinus size={20} />
                      </button>
                    </div>
                  ))
                )}
              </div>

              {/* Friends Pagination */}
              {Math.ceil(friendsTotalCount / pageSize) > 1 && (
                <div className="mt-12 flex justify-center items-center gap-6">
                  <button
                    onClick={() => setFriendsPageNumber(p => Math.max(1, p - 1))}
                    disabled={friendsPageNumber === 1 || loading}
                    className="px-6 py-2 rounded-xl bg-white/5 border border-white/10 text-white font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                  >
                    PREVIOUS
                  </button>
                  <div className="flex flex-col items-center">
                    <span className="text-white font-black text-xl tracking-tighter">{friendsPageNumber}</span>
                    <span className="text-gray-500 text-[10px] font-bold uppercase tracking-widest">OF {Math.ceil(friendsTotalCount / pageSize)}</span>
                  </div>
                  <button
                    onClick={() => setFriendsPageNumber(p => Math.min(Math.ceil(friendsTotalCount / pageSize), p + 1))}
                    disabled={friendsPageNumber === Math.ceil(friendsTotalCount / pageSize) || loading}
                    className="px-6 py-2 rounded-xl bg-white/5 border border-white/10 text-white font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                  >
                    NEXT
                  </button>
                </div>
              )}
            </>
          ) : (
            <>
              <div className="flex justify-between items-start">
                <div>
                  <h2 className="text-3xl font-black text-white tracking-tight">Pending Requests</h2>
                  <p className="text-gray-500 text-sm font-bold uppercase tracking-widest">People who want to be your friend</p>
                </div>
              </div>

              <div className="space-y-4">
                {invitations.length === 0 ? (
                  <div className="text-center py-20 text-gray-500 font-bold italic">
                    No pending invitations.
                  </div>
                ) : (
                  invitations.map((invitation) => (
                    <div key={invitation.id} className="flex items-center justify-between p-4 glass-light rounded-2xl border border-white/5">
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-full bg-secondary/20 flex items-center justify-center text-secondary font-bold ring-1 ring-secondary/30">
                          {invitation.sender?.name.slice(0, 2).toUpperCase()}
                        </div>
                        <div>
                          <h3 className="text-white font-bold text-lg">{invitation.sender?.name}</h3>
                          <p className="text-gray-500 text-sm font-mono tracking-tighter">#{invitation.sender?.identifier}</p>
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <button
                          onClick={() => handleRespond(invitation.id, true)}
                          className="p-3 bg-primary/20 hover:bg-primary text-primary hover:text-background rounded-xl transition-all"
                          title="Accept"
                        >
                          <Check size={20} />
                        </button>
                        <button
                          onClick={() => handleRespond(invitation.id, false)}
                          className="p-3 bg-red-500/10 hover:bg-red-500 text-red-500 hover:text-white rounded-xl transition-all"
                          title="Decline"
                        >
                          <X size={20} />
                        </button>
                      </div>
                    </div>
                  ))
                )}
              </div>

              {/* Invitations Pagination */}
              {Math.ceil(invitationsTotalCount / pageSize) > 1 && (
                <div className="mt-12 flex justify-center items-center gap-6">
                  <button
                    onClick={() => setInvitationsPageNumber(p => Math.max(1, p - 1))}
                    disabled={invitationsPageNumber === 1 || loading}
                    className="px-6 py-2 rounded-xl bg-white/5 border border-white/10 text-white font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                  >
                    PREVIOUS
                  </button>
                  <div className="flex flex-col items-center">
                    <span className="text-white font-black text-xl tracking-tighter">{invitationsPageNumber}</span>
                    <span className="text-gray-500 text-[10px] font-bold uppercase tracking-widest">OF {Math.ceil(invitationsTotalCount / pageSize)}</span>
                  </div>
                  <button
                    onClick={() => setInvitationsPageNumber(p => Math.min(Math.ceil(invitationsTotalCount / pageSize), p + 1))}
                    disabled={invitationsPageNumber === Math.ceil(invitationsTotalCount / pageSize) || loading}
                    className="px-6 py-2 rounded-xl bg-white/5 border border-white/10 text-white font-bold disabled:opacity-30 disabled:cursor-not-allowed hover:bg-white/10 transition-colors"
                  >
                    NEXT
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      <CreationDialog
        isOpen={isInviteDialogOpen}
        onClose={() => setIsInviteDialogOpen(false)}
        onCreate={handleInvite}
        title="Invite Friend"
        placeholder="Enter identifier (e.g. Name#12345678)"
        label="Friend Identifier"
        submitLabel="Send Invitation"
      />
    </div>
  );
};
