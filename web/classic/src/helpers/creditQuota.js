/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { getQuotaPerUnit } from './quota';

const defaultCreditDivisor = 5000;

export const getCreditDivisor = () => {
  const raw = parseFloat(localStorage.getItem('credit_divisor') || '');
  if (Number.isFinite(raw) && raw > 0) {
    return raw;
  }

  const quotaPerUnit = getQuotaPerUnit();
  if (Number.isFinite(quotaPerUnit) && quotaPerUnit > 0) {
    return quotaPerUnit / 100;
  }

  return defaultCreditDivisor;
};

export const quotaToCredits = (quota) => {
  const value = Number(quota || 0);
  if (!Number.isFinite(value) || value === 0) {
    return 0;
  }
  return value / getCreditDivisor();
};

export const creditsToQuota = (credits) => {
  const value = Number(credits || 0);
  if (!Number.isFinite(value) || value === 0) {
    return 0;
  }
  return Math.round(value * getCreditDivisor());
};

export const formatCredits = (credits, digits = 6) => {
  const value = Number(credits || 0);
  if (!Number.isFinite(value)) {
    return '0';
  }
  return value
    .toFixed(digits)
    .replace(/(\.\d*?[1-9])0+$/, '$1')
    .replace(/\.0+$/, '');
};
