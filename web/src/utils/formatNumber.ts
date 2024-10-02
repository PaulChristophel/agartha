import numeral from 'numeral';

// ----------------------------------------------------------------------

export function fNumber(number: number): string {
  return numeral(number).format();
}

export function fCurrency(number: number): string {
  const format = number ? numeral(number).format('$0,0.00') : '';
  return result(format, '.00');
}

export function fPercent(number: number): string {
  const format = number ? numeral(Number(number) / 100).format('0.0%') : '';
  return result(format, '.0');
}

export function fShortenNumber(number: number): string {
  const format = number ? numeral(number).format('0.00a') : '';
  return result(format, '.00');
}

export function fData(number: number): string {
  const format = number ? numeral(number).format('0.0 b') : '';
  return result(format, '.0');
}

function result(format: string, key: string = '.00'): string {
  const isInteger = format.includes(key);
  return isInteger ? format.replace(key, '') : format;
}
