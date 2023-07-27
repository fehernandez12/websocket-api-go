function isPrime(num: number): boolean {
  return !"1".repeat(num).match(/^1?$|^(11+?)\1+$/);
}
