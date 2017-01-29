import { ClassifyItem } from '../../classify/classify.service'

export interface Movie {
	name          : string
	url           : string
	released      : string
	duration      : number
	image         : string
	description   : string
	directors     : string[]
	cast          : string[]
	genres        : string[]
}
