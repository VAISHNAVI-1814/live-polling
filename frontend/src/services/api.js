const API_BASE = import.meta.env.VITE_API_URL || `${typeof window !== 'undefined' ? window.location.protocol : 'http:'}//${typeof window !== 'undefined' ? window.location.hostname : 'localhost'}:8080/api`;
const WS_BASE = import.meta.env.VITE_WS_URL || `${typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${typeof window !== 'undefined' ? window.location.hostname : 'localhost'}:8080/ws`;


// Generate or retrieve persistent voter token from localStorage
export function getVoterToken() {
  let token = localStorage.getItem('poll_voter_token');
  if (!token) {
    token = 'voter_' + Math.random().toString(36).substring(2, 15) + '_' + Date.now().toString(36);
    localStorage.setItem('poll_voter_token', token);
  }
  return token;
}

// Helper for HTTP requests
async function request(endpoint, options = {}) {
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  const token = localStorage.getItem('auth_token');
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  const contentType = res.headers.get('content-type');
  let data = null;
  if (contentType && contentType.includes('application/json')) {
    data = await res.json();
  }

  if (!res.ok) {
    const errorMsg = data?.error || `HTTP error ${res.status}`;
    const err = new Error(errorMsg);
    err.status = res.status;
    err.data = data;
    throw err;
  }

  return data;
}

export const api = {
  // Auth
  async signup(name, email, password) {
    return request('/auth/signup', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    });
  },

  async login(email, password) {
    return request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  },

  async getMe() {
    return request('/auth/me');
  },

  // Polls
  async createPoll(payload) {
    return request('/polls', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async getMyPolls() {
    return request('/polls/my-polls');
  },

  async getPollById(id) {
    return request(`/polls/${id}`);
  },

  async getPollByShareCode(code) {
    return request(`/polls/share/${code}`);
  },

  async vote(pollId, optionId) {
    const voterToken = getVoterToken();
    return request(`/polls/${pollId}/vote`, {
      method: 'POST',
      body: JSON.stringify({ option_id: optionId, voter_token: voterToken }),
    });
  },

  async getResults(pollId) {
    return request(`/polls/${pollId}/results`);
  },

  async checkIfVoted(pollId) {
    const voterToken = getVoterToken();
    return request(`/polls/${pollId}/voted?voterToken=${encodeURIComponent(voterToken)}`);
  },

  async updatePollStatus(pollId, status) {
    return request(`/polls/${pollId}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  },

  async deletePoll(pollId) {
    return request(`/polls/${pollId}`, {
      method: 'DELETE',
    });
  },

  getWebSocketUrl(pollId) {
    return `${WS_BASE}/polls/${pollId}`;
  },
};
