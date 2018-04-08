import { NgModule } from '@angular/core'
import { CommonModule } from '@angular/common'
import { BrowserModule } from '@angular/platform-browser'
import { FormsModule } from '@angular/forms'
import { ConfigsModule } from '../configs/configs.module'
import { ToolsModule } from '../tools/tools.module'
import { ParamsModule } from '../params/params.module'

import { ImportsService } from './imports.service'

import { ImportsComponent } from './imports.component'
import { ImportCreateComponent } from './create.component'
import { DirectoryCreateComponent } from './directory/create.component'
import { DirectoryDisplayComponent } from './directory/display.component'
import { ImapCreateComponent } from './imap/create.component'
import { ImapSearchComponent } from './imap/search.component'
import { ImapDisplayComponent } from './imap/display.component'

@NgModule({
    imports: [
        CommonModule,
        BrowserModule,
        FormsModule,
		ConfigsModule,
        ToolsModule,
        ParamsModule
    ],
    providers: [ImportsService],
    declarations: [ImportsComponent,
        DirectoryCreateComponent,
        DirectoryDisplayComponent,
        ImapCreateComponent,
        ImapSearchComponent,
        ImapDisplayComponent],
    exports: [ImportsComponent]
})

export class ImportsModule { }
