import React, { useState } from 'react';
import { X, Copy, Check, ExternalLink, Share2 } from 'lucide-react';

export default function ShareModal({ poll, isOpen, onClose }) {
  const [copied, setCopied] = useState(false);

  if (!isOpen || !poll) return null;

  const origin = window.location.origin;
  const shareUrl = `${origin}/poll/${poll.share_code}`;
  const resultsUrl = `${origin}/poll/${poll.id}/results`;

  const copyToClipboard = (url) => {
    navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fadeIn">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl relative">
        <button
          onClick={onClose}
          className="absolute top-4 right-4 p-1.5 text-slate-400 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
        >
          <X className="w-5 h-5" />
        </button>

        <div className="flex items-center gap-3 mb-4">
          <div className="p-2.5 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            <Share2 className="w-6 h-6" />
          </div>
          <div>
            <h3 className="font-semibold text-white text-lg">Share Poll</h3>
            <p className="text-xs text-slate-400">Share this code or link with your audience to vote live.</p>
          </div>
        </div>

        {/* Share Code Badge */}
        <div className="bg-slate-950 border border-slate-800 rounded-xl p-4 text-center mb-5">
          <span className="text-xs text-slate-400 uppercase tracking-wider block mb-1">Share Code</span>
          <span className="text-3xl font-extrabold tracking-widest text-indigo-400 font-mono">
            {poll.share_code}
          </span>
        </div>

        {/* Voting Link */}
        <div className="mb-4">
          <label className="block text-xs font-medium text-slate-300 mb-1.5">Audience Voting Link</label>
          <div className="flex items-center gap-2">
            <input
              type="text"
              readOnly
              value={shareUrl}
              className="flex-1 bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-300 font-mono focus:outline-none focus:border-indigo-500 select-all"
            />
            <button
              onClick={() => copyToClipboard(shareUrl)}
              className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors flex-shrink-0"
            >
              {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
        </div>

        {/* Live Results Link */}
        <div className="mb-6">
          <label className="block text-xs font-medium text-slate-300 mb-1.5">Live Results Link (Projector / Screen)</label>
          <div className="flex items-center gap-2">
            <input
              type="text"
              readOnly
              value={resultsUrl}
              className="flex-1 bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-slate-300 font-mono focus:outline-none focus:border-indigo-500 select-all"
            />
            <a
              href={resultsUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg transition-colors flex-shrink-0"
            >
              <ExternalLink className="w-3.5 h-3.5" />
              Open
            </a>
          </div>
        </div>

        <div className="text-center">
          <button
            onClick={onClose}
            className="w-full py-2.5 text-sm font-medium text-slate-300 bg-slate-800 hover:bg-slate-700 rounded-xl transition-colors"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  );
}
