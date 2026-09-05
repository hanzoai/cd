import {DataLoader} from '../../../../kit/src';
import * as models from '../../../shared/models';

export type LogLoader = DataLoader<models.LogEntry[], string>;
