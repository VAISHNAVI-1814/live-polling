import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../services/api';
import { usePollWebSocket } from '../hooks/usePollWebSocket';
import LiveResultBar from '../components/LiveResultBar';
import ShareModal from '../components/ShareModal';
import { 
  BarChart3, 
  Radio, 
  Share2, 
  ExternalLink, 
  Clock, 
  AlertCircle, 
  Users, 
  Award,
  Vote
} from 'lucide-react';

export default function ResultsPage() {
  const { id } = useParams();
  const [results, setResults] = useState(null);
  const [poll, setPoll] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [isShareModalOpen, setIsShareModalOpen] = useState(false);

  // Hook up WebSocket for real-time live push updates
  const { connectionStatus, lastMessageTime } = usePollWebSocket(id, (liveUpdate) => {
    if (liveUpdate) {
      setResults(liveUpdate);
    }
  });

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true);
        setError('');
        const [resultsData, pollData] = await Promise.all([
          api.getResults(id),
          api.getPollById(id).catch(() => null),
        ]);
        setResults(resultsData);
        setPoll(pollData);
      } catch (err) {
        setError(err.message || 'Failed to load live results');
      } finally {
        setLoading(false);
      }
    }

    if (id) {
      loadData();
    }
  }, [id]);

  if (loading) {
    return (
      <div className="min-h-[80vh] flex items-center justify-center">
        <div className="animate-spin rounded-full h-10 w-10 border-t-2 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  if (error || !results) {
    return (
      <div className="max-w-md mx-auto px-4 py-20 text-center">
        <div className="w-12 h-12 rounded-2xl bg-rose-500/10 text-rose-400 flex items-center justify-center mx-auto mb-4 border border-rose-500/20">
          <AlertCircle className="w-6 h-6" />
        </div>
        <h2 className="text-xl font-bold text-white mb-2">Results Unavailable</h2>
        <p className="text-sm text-slate-400 mb-6">{error || 'Poll not found'}</p>
        <Link
          to="/dashboard"
          className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-500 rounded-xl transition-colors"
        >
          Return to Dashboard
        </Link>
      </div>
    );
  }

  const maxVotes = Math.max(...(results.options || []).map(o => o.votes || 0), 0);
  const isClosed = results.status === 'closed';
  const isExpired = results.is_expired;

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      {/* Top Controls & Status Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
        <div className="flex items-center gap-3">
          {/* Live Pulsing WebSocket Status Indicator */}
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-slate-900 border border-slate-800 text-xs">
            <span className="flex h-2 w-2 relative">
              {connectionStatus === 'connected' ? (
                <>
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                </>
              ) : connectionStatus === 'connecting' ? (
                <span className="relative inline-flex rounded-full h-2 w-2 bg-amber-500 animate-pulse"></span>
              ) : (
                <span className="relative inline-flex rounded-full h-2 w-2 bg-slate-500"></span>
              )}
            </span>
            <span className="font-medium text-slate-300">
              {connectionStatus === 'connected' ? 'Live Stream Active' : connectionStatus === 'connecting' ? 'Connecting...' : 'Offline'}
            </span>
          </div>

          <span className="font-mono text-xs px-3 py-1.5 rounded-full bg-slate-900 text-indigo-400 border border-slate-800 font-semibold">
            Code: {results.share_code}
          </span>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            onClick={() => setIsShareModalOpen(true)}
            className="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-xl transition-colors"
          >
            <Share2 className="w-3.5 h-3.5 text-indigo-400" />
            Share Link
          </button>

          <Link
            to={`/poll/${results.share_code}`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 px-3.5 py-2 text-xs font-semibold bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 border border-indigo-500/30 rounded-xl transition-colors"
          >
            <Vote className="w-3.5 h-3.5" />
            Open Voting Tab
            <ExternalLink className="w-3 h-3" />
          </Link>
        </div>
      </div>

      {/* Main Results Card */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-3xl p-6 sm:p-10 shadow-2xl backdrop-blur-sm relative overflow-hidden">
        {/* Glow Accent Header */}
        <div className="absolute top-0 left-0 right-0 h-1.5 bg-gradient-to-r from-indigo-500 via-purple-500 to-emerald-500" />

        {/* Question Header */}
        <div className="mb-8">
          <div className="flex items-center gap-2 text-xs text-slate-400 mb-2">
            <span className={`px-2.5 py-0.5 rounded-full font-semibold ${
              !isClosed && !isExpired
                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                : 'bg-slate-800 text-slate-400 border border-slate-700'
            }`}>
              {isExpired ? 'Expired' : isClosed ? 'Closed' : 'Active'}
            </span>
            {results.expires_at && (
              <span className="flex items-center gap-1">
                <Clock className="w-3.5 h-3.5 text-slate-500" />
                {isExpired ? 'Expired' : `Closes at ${new Date(results.expires_at).toLocaleTimeString()}`}
              </span>
            )}
          </div>

          <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight leading-snug">
            {results.question}
          </h1>
        </div>

        {/* Metric Summary */}
        <div className="flex items-center justify-between p-4 rounded-2xl bg-slate-950 border border-slate-800/80 mb-8">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-indigo-500/10 text-indigo-400">
              <Users className="w-5 h-5" />
            </div>
            <div>
              <span className="text-xs text-slate-400 block">Total Votes Cast</span>
              <span className="text-2xl font-bold text-white font-mono">
                {results.total_votes || 0}
              </span>
            </div>
          </div>

          {lastMessageTime && (
            <div className="text-right text-xs text-slate-500 hidden sm:block">
              <span>Last live update</span>
              <span className="block font-mono text-slate-400">
                {lastMessageTime.toLocaleTimeString()}
              </span>
            </div>
          )}
        </div>

        {/* Live Option Bars */}
        <div className="space-y-4">
          {results.options.map((opt) => {
            const isLeader = maxVotes > 0 && opt.votes === maxVotes;
            return (
              <LiveResultBar
                key={opt.option_id}
                option={opt}
                isLeader={isLeader}
                totalVotes={results.total_votes || 0}
              />
            );
          })}
        </div>

        {/* Real-time Hint Footer */}
        <div className="mt-8 pt-6 border-t border-slate-800 flex flex-col sm:flex-row items-center justify-between text-xs text-slate-500 gap-2">
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-emerald-400 animate-pulse" />
            <span>Powered by Redis Pub/Sub & WebSocket real-time engine.</span>
          </div>
          <span className="font-medium text-slate-400">
            No page refresh required.
          </span>
        </div>
      </div>

      {/* Share Modal */}
      <ShareModal
        poll={poll || { id: results.poll_id, share_code: results.share_code }}
        isOpen={isShareModalOpen}
        onClose={() => setIsShareModalOpen(false)}
      />
    </div>
  );
}
