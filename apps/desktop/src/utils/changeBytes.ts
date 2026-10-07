export const changeBytes = (bytes: number, decimals = 2) => {
  if (bytes === 0) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
};

export function getStorageUsage(totalStorage: number, usedStorage: number) {
  if (totalStorage <= 0) {
    return {
      usedPercentage: 0,
      remainingPercentage: 0,
    };
  }

  const usedPercentage = (usedStorage / totalStorage) * 100;
  const remainingPercentage = 100 - usedPercentage;

  return {
    usedPercentage: +usedPercentage.toFixed(2),
    remainingPercentage: +remainingPercentage.toFixed(2),
  };
}
