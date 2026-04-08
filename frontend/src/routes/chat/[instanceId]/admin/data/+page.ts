export const load = ({ url }) => ({
  selectedStream: url.searchParams.get('stream'),
  selectedKV: url.searchParams.get('kv')
});
