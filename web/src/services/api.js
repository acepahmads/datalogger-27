import axios from 'axios';

const api = {
  get(url, config) {
    const fullUrl = url.startsWith('/api') ? url : `/api${url.startsWith('/') ? url : '/' + url}`;
    return axios.get(fullUrl, config);
  },
  post(url, data, config) {
    const fullUrl = url.startsWith('/api') ? url : `/api${url.startsWith('/') ? url : '/' + url}`;
    return axios.post(fullUrl, data, config);
  },
  put(url, data, config) {
    const fullUrl = url.startsWith('/api') ? url : `/api${url.startsWith('/') ? url : '/' + url}`;
    return axios.put(fullUrl, data, config);
  },
  delete(url, config) {
    const fullUrl = url.startsWith('/api') ? url : `/api${url.startsWith('/') ? url : '/' + url}`;
    return axios.delete(fullUrl, config);
  },
  patch(url, data, config) {
    const fullUrl = url.startsWith('/api') ? url : `/api${url.startsWith('/') ? url : '/' + url}`;
    return axios.patch(fullUrl, data, config);
  },
};

export default api;
