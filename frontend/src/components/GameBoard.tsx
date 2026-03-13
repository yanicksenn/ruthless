import React, { useState, useEffect } from 'react';
import { gameClient, sessionClient, deckClient, createOptions } from '../api/client';
import { Game, GameState, Card, Session } from '../api/ruthless';
import { useAuth } from '../context/AuthContext';
import { ArrowLeft, Play, Crown, Check, Info, Users, Layers } from 'lucide-react';
import { motion } from 'framer-motion';

interface GameBoardProps {
  sessionId: string;
  onLeave: () => void;
}

export const GameBoard: React.FC<GameBoardProps> = ({ sessionId, onLeave }) => {
  const { token, user } = useAuth();
  const [session, setSession] = useState<Session | null>(null);
  const [game, setGame] = useState<Game | null>(null);
  const [hand, setHand] = useState<Card[]>([]);
  const [selectedCards, setSelectedCards] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [decks, setDecks] = useState<{id: string, name: string}[]>([]);

  const isOwner = session?.ownerId === user?.id;

  const fetchData = async () => {
    try {
      const sResponse = await sessionClient.getSession({ id: sessionId }, createOptions(token));
      setSession(sResponse.response);

      try {
        const gResponse = await gameClient.getGameBySession({ sessionId }, createOptions(token));
        setGame(gResponse.response);
        
        if (gResponse.response.state === GameState.PLAYING) {
           const hResponse = await gameClient.getHand({ gameId: gResponse.response.id }, createOptions(token));
           setHand(hResponse.response.cards);
        }
      } catch (gErr) {
        // Game might not be created yet, that's fine
        setGame(null);
      }

      if (isOwner && !game) {
        const dResponse = await deckClient.listDecks({}, createOptions(token));
        setDecks(dResponse.response.decks.map(d => ({ id: d.id, name: d.name })));
      }
    } catch (err) {
      console.error('Fetch error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 3000);
    return () => clearInterval(interval);
  }, [sessionId, token, user?.id]);

  const handleCreateGame = async () => {
    try {
      await gameClient.createGame({ sessionId }, createOptions(token));
      fetchData();
    } catch (err) { console.error(err); }
  };

  const handleStartGame = async () => {
    if (!game) return;
    try {
      await gameClient.startGame({ id: game.id }, createOptions(token));
      fetchData();
    } catch (err) { console.error(err); }
  };

  const handleAddDeck = async (deckId: string) => {
    try {
       await sessionClient.addDeckToSession({ sessionId, deckId }, createOptions(token));
       fetchData();
    } catch (err) { console.error(err); }
  };

  const handlePlayCards = async () => {
    if (!game) return;
    try {
      await gameClient.playCards({ gameId: game.id, cardIds: selectedCards }, createOptions(token));
      setSelectedCards([]);
      fetchData();
    } catch (err) { console.error(err); }
  };

  const handleSelectWinner = async (playId: string) => {
    if (!game) return;
    try {
      await gameClient.selectWinner({ gameId: game.id, playId }, createOptions(token));
      fetchData();
    } catch (err) { console.error(err); }
  };

  if (loading && !session) return <div className="flex items-center justify-center min-h-screen">Loading...</div>;

  const currentRound = game?.rounds[game.rounds.length - 1];
  const isCzar = currentRound?.czarId === user?.id;

  return (
    <div className="min-h-screen flex flex-col">
      {/* Top Bar */}
      <nav className="glass border-b border-white/5 px-6 py-4 flex justify-between items-center sticky top-0 z-50">
        <button onClick={onLeave} className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors">
          <ArrowLeft size={18} /> <span className="font-bold text-xs uppercase tracking-widest">Quit to Lobby</span>
        </button>
        <div className="flex items-center gap-6">
           <div className="flex items-center gap-2">
             <Users size={16} className="text-primary" />
             <span className="font-bold text-sm tracking-tight">{session?.playerIds.length} Players</span>
           </div>
           {game && (
             <div className="bg-white/5 px-3 py-1 rounded-full text-[10px] font-black uppercase text-primary border border-primary/20">
               {GameState[game.state].replace('GAME_STATE_', '')}
             </div>
           )}
        </div>
      </nav>

      <main className="flex-1 p-6 md:p-12">
        {!game ? (
          <div className="max-w-2xl mx-auto space-y-8 mt-12">
             <div className="text-center">
               <h2 className="text-4xl font-black mb-2 tracking-tighter uppercase">Preparation</h2>
               <p className="text-gray-400 font-bold tracking-widest uppercase text-xs italic">Waiting for the host to summon the cards</p>
             </div>
             
             {isOwner ? (
                <div className="glass p-8 rounded-3xl space-y-8 card-shadow">
                   <div className="space-y-4">
                     <h3 className="flex items-center gap-2 font-black uppercase text-sm tracking-widest text-primary">
                       <Layers size={18} /> Select Decks ({session?.deckIds.length || 0})
                     </h3>
                     <div className="grid grid-cols-2 gap-2">
                        {decks.map(d => (
                          <button 
                            key={d.id} 
                            onClick={() => handleAddDeck(d.id)}
                            disabled={session?.deckIds.includes(d.id)}
                            className={`p-3 rounded-xl border text-sm font-bold transition-all text-left flex justify-between items-center ${
                              session?.deckIds.includes(d.id) 
                              ? 'bg-primary/20 border-primary/40 text-primary' 
                              : 'bg-white/5 border-white/10 hover:border-white/20'
                            }`}
                          >
                            {d.name}
                            {session?.deckIds.includes(d.id) && <Check size={14} />}
                          </button>
                        ))}
                     </div>
                   </div>
                   
                   <button 
                     onClick={handleCreateGame} 
                     disabled={!session?.deckIds.length}
                     className="w-full bg-primary hover:bg-primary-dark disabled:opacity-50 text-white font-black py-4 rounded-2xl flex items-center justify-center gap-2 transition-transform active:scale-95"
                   >
                     <Play size={20} fill="currentColor" /> CREATE GAME
                   </button>
                </div>
             ) : (
                <div className="glass p-12 rounded-3xl text-center border-dashed border-2 border-white/10">
                   <div className="animate-pulse flex flex-col items-center gap-4">
                      <div className="w-16 h-16 bg-white/5 rounded-full flex items-center justify-center border border-white/10">
                        <Users size={32} className="text-gray-500" />
                      </div>
                      <p className="text-gray-400 font-bold uppercase tracking-widest text-xs">Waiting for {session?.ownerId.split('-')[0]}...</p>
                   </div>
                </div>
             )}
          </div>
        ) : game.state === GameState.WAITING ? (
           <div className="max-w-md mx-auto mt-24 text-center space-y-8">
              <div className="glass p-12 rounded-3xl space-y-8 border-primary/20 border-2">
                 <div className="flex justify-center">
                    <div className="bg-primary/20 p-4 rounded-full border border-primary/40 text-primary">
                       <Play size={40} fill="currentColor" />
                    </div>
                 </div>
                 <div>
                    <h2 className="text-3xl font-black mb-2">READY TO START?</h2>
                    <p className="text-gray-400 text-sm font-bold uppercase tracking-wider">All players accounted for</p>
                 </div>
                 {isOwner ? (
                    <button onClick={handleStartGame} className="w-full bg-primary hover:bg-primary-dark text-white font-black py-4 rounded-2xl">
                      START FIRST ROUND
                    </button>
                 ) : (
                    <p className="text-primary font-bold animate-bounce">Host is about to start...</p>
                 )}
              </div>
           </div>
        ) : (
           <div className="space-y-12">
              {/* Black Card Display */}
              <div className="flex justify-center">
                 <motion.div 
                   layoutId="black-card"
                   className="bg-black border border-white/20 p-8 rounded-2xl w-full max-w-sm aspect-[3/4] flex flex-col justify-between card-shadow shadow-primary/20"
                 >
                    <p className="text-2xl font-black leading-tight tracking-tight">
                       {currentRound?.blackCard?.text || "..."}
                    </p>
                    <div className="flex justify-between items-end">
                       <div className="text-[10px] font-black tracking-[0.2em] opacity-30">RUTHLESS</div>
                       <Crown className={isCzar ? "text-primary fill-primary" : "text-gray-800"} size={24} />
                    </div>
                 </motion.div>
              </div>

              {/* Status Indicator */}
              <div className="text-center space-y-2">
                 {isCzar ? (
                   <div className="inline-flex items-center gap-2 bg-primary/10 border border-primary/20 px-4 py-1.5 rounded-full text-primary font-black text-[10px] uppercase tracking-widest">
                      <Crown size={12} fill="currentColor" /> You are the Czar
                   </div>
                 ) : (
                   <div className="inline-flex items-center gap-2 bg-white/5 border border-white/10 px-4 py-1.5 rounded-full text-gray-400 font-black text-[10px] uppercase tracking-widest">
                      <Users size={12} /> Submit your play
                   </div>
                 )}
                 <h3 className="text-xl font-bold italic text-gray-500">
                    {game.state === GameState.PLAYING ? "Players are choosing..." : "Czar is judging..."}
                 </h3>
              </div>

              {/* Submissions or Hand */}
              <div className="mt-12">
                 {game.state === GameState.JUDGING ? (
                   <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                      {Object.entries(currentRound?.plays || {}).map(([pid, play]) => (
                        <div 
                           key={pid} 
                           className={`glass p-6 rounded-2xl flex flex-col justify-between aspect-[4/3] relative group transition-all ${isCzar ? 'hover:border-primary cursor-pointer' : ''}`}
                           onClick={() => isCzar && handleSelectWinner(play.id)}
                        >
                           <div className="space-y-2">
                              {play.cards.map(c => (
                                <p key={c.id} className="text-lg font-bold">{c.text}</p>
                              ))}
                           </div>
                           {isCzar && (
                             <div className="absolute inset-0 bg-primary/10 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded-2xl">
                                <span className="bg-primary text-white font-black px-4 py-2 rounded-lg text-sm">ELECT WINNER</span>
                             </div>
                           )}
                        </div>
                      ))}
                   </div>
                 ) : !isCzar ? (
                   <div className="space-y-6">
                      <div className="flex justify-between items-center">
                         <h3 className="font-black text-xs uppercase tracking-[0.2em] text-gray-500">Your Hand</h3>
                         <button 
                           onClick={handlePlayCards}
                           disabled={selectedCards.length === 0}
                           className="bg-primary hover:bg-primary-dark disabled:opacity-30 px-6 py-2 rounded-xl text-xs font-black transition-all"
                         >
                           SUBMIT SELECTION
                         </button>
                      </div>
                      <div className="grid grid-cols-2 md:grid-cols-5 gap-3">
                         {hand.map((card) => (
                           <div 
                              key={card.id}
                              onClick={() => {
                                if (selectedCards.includes(card.id)) setSelectedCards(s => s.filter(id => id !== card.id));
                                else setSelectedCards(s => [...s, card.id]);
                              }}
                              className={`p-4 rounded-xl border aspect-[1/1] text-sm font-bold transition-all flex flex-col justify-between cursor-pointer ${
                                selectedCards.includes(card.id) 
                                ? 'bg-white text-black border-white scale-105 z-10' 
                                : 'bg-surface border-white/5 text-gray-300 hover:border-white/20'
                              }`}
                           >
                              <p className="line-clamp-4">{card.text}</p>
                              <div className="flex justify-end opacity-20"><Info size={12} /></div>
                           </div>
                         ))}
                      </div>
                   </div>
                 ) : (
                    <div className="bg-white/5 border border-dashed border-white/10 p-12 rounded-3xl text-center">
                       <p className="text-gray-500 font-bold italic">Wait for the plebs to finish their selections...</p>
                    </div>
                 )}
              </div>
           </div>
        )}
      </main>
    </div>
  );
};
