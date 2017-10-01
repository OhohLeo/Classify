import { NgModule } from '@angular/core'
import { CommonModule } from '@angular/common'
import { BrowserModule } from '@angular/platform-browser'
import { FormsModule } from '@angular/forms'
import { ToolsModule } from '../tools/tools.module'

import { ConfigsComponent } from './configs.component'
import { ConfigsService } from './configs.service'

@NgModule({
    imports: [
        CommonModule,
        BrowserModule,
        FormsModule,
        ToolsModule
    ],
    declarations: [ConfigsComponent],
    providers: [ConfigsService],
    exports: [ConfigsComponent]
})

export class ConfigsModule { }